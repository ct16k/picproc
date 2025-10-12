package mangle

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"picproc/palette"
	"picproc/parallel"

	"github.com/alecthomas/kong"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

type CLICmd struct {
	Scan      string      `help:"Source folder to scan" default:"."`
	Dest      string      `help:"Destination folder for processed pictures. Relative to scan dir if not absolute. If same as scan dir, will overwrite source files." default:"mangled"`
	Resize    bool        `help:"Resize image" default:"false" group:"resize"`
	Width     int         `help:"Max width" group:"resize"`
	Height    int         `help:"Max height" group:"resize"`
	Crop      bool        `help:"Crop image to maintain requested aspect ration" default:"false" group:"resize"`
	Fill      string      `help:"If given and not cropping, will fill background with this color to maintain destination aspect ratio" group:"resize"`
	Palette   string      `help:"Palette name (bw, spectra6, mattdm6, gray16, vga16, vga256) or PAL file in RIFF format to apply" group:"palette"`
	Dither    bool        `help:"Apply dithering" default:"false" group:"palette"`
	Format    string      `help:"Output format of mangled image. If prefixed with 'unsup:' will convert only unsupported formats" enum:"same,gif,unsup:gif,jpeg,unsup:jpeg,png,unsup:png,bmp,unsup:bmp,tiff,unsup:tiff" default:"unsup:png"`
	FillColor color.Color `kong:"-"`
}

func (c *CLICmd) Validate(kctx *kong.Context) error {
	scanDir, err := filepath.Abs(c.Scan)
	var info os.FileInfo
	if err == nil {
		if info, err = os.Stat(scanDir); err == nil && !info.IsDir() {
			err = fmt.Errorf("not a directory")
		}
	}
	if err != nil {
		return fmt.Errorf("invalid scan path %q: %w", c.Scan, err)
	}
	c.Scan = scanDir

	if !filepath.IsAbs(c.Dest) {
		c.Dest = filepath.Join(scanDir, c.Dest)
	}

	if c.Resize {
		switch {
		case (c.Width < 0):
			return fmt.Errorf("invalid resize width: %d", c.Width)
		case (c.Height < 0):
			return fmt.Errorf("invalid resize height: %d", c.Height)
		case (c.Width == 0) && (c.Height == 0):
			return fmt.Errorf("no resize dimensions given")
		}
	}

	if (!c.Crop) && (c.Fill != "") {
		if c.FillColor, err = parseHexToColor(c.Fill); err != nil {
			return err
		}
	}

	if c.Palette != "" {
		if _, err := palette.LoadPalette(c.Palette); err != nil {
			return err
		}
	}

	return nil
}

func (c *CLICmd) Run(worker parallel.WorkerFunc, wait parallel.WaitFunc) error {
	if err := os.MkdirAll(c.Dest, os.ModeDir); err != nil {
		return fmt.Errorf("unable to create destination folder %q: %w", c.Dest, err)
	}

	files, err := os.ReadDir(c.Scan)
	if err != nil {
		return fmt.Errorf("unable to read folder %q: %w", c.Scan, err)
	}

	var processedCount, errCount atomic.Uint64
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		worker(func(fileName string) func() {
			return func() {
				filePath := filepath.Join(c.Scan, fileName)
				logger := slog.Default().With("file", filePath)

				imgFile, err := os.Open(filePath)
				if err != nil {
					errCount.Add(1)
					logger.Error("could not open image", "error", err)
					return
				}

				img, imgType, err := image.Decode(imgFile)
				if err != nil {
					errCount.Add(1)
					logger.Error("could not decode image", "error", err)
					return
				}

				if c.Resize {
					img, err = resize(logger, img, c.Width, c.Height, c.Crop, c.FillColor)
					if err != nil {
						errCount.Add(1)
						logger.Error("could not resize image", "error", err)
						return
					}
				}

				if c.Palette != "" {
					palLog := logger.With("palette", c.Palette)
					img, err = repallete(palLog, img, c.Palette, c.Dither)
					if err != nil {
						errCount.Add(1)
						palLog.Error("could not change image pallete", "error", err)
						return
					}
				}

				if err = save(img, imgType, c.Format, c.Dest, fileName); err != nil {
					errCount.Add(1)
					logger.Error("could not save image", "dir", c.Dest, "error", err)
					return
				}
				processedCount.Add(1)
			}
		}(file.Name()))
	}

	wait(true)

	processed := processedCount.Load()
	errors := errCount.Load()
	slog.Info("stats", "processed", processed, "errors", errors,
		"total", processed+errors)

	if errors > 0 {
		return fmt.Errorf("error processing %d files", errors)
	}
	return nil
}

func parseHexToColor(s string) (color.Color, error) {
	var c color.RGBA
	switch len(s) {
	case 4:
		n, err := fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		if err != nil {
			return nil, fmt.Errorf("could not read color: %w", err)
		} else if n < 3 {
			return nil, fmt.Errorf("insufficient fill color fields: %d", n)
		}

		c.R |= c.R << 4
		c.G |= c.G << 4
		c.B |= c.B << 4
		c.A = 0xFF
	case 5:
		n, err := fmt.Sscanf(s, "#%1x%1x%1x%x", &c.R, &c.G, &c.B, &c.A)
		if err != nil {
			return nil, fmt.Errorf("could not read color: %w", err)
		} else if n < 3 {
			return nil, fmt.Errorf("insufficient fill color fields: %d", n)
		}

		c.R |= c.R << 4
		c.G |= c.G << 4
		c.B |= c.B << 4
		c.A |= c.A << 4
	case 7:
		n, err := fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		if err != nil {
			return nil, fmt.Errorf("could not read color: %w", err)
		} else if n < 3 {
			return nil, fmt.Errorf("insufficient fill color fields: %d", n)
		}

		c.A = 0xFF
	case 8:
		n, err := fmt.Sscanf(s, "#%1x%1x%1x%x", &c.R, &c.G, &c.B, &c.A)
		if err != nil {
			return nil, fmt.Errorf("could not read color: %w", err)
		} else if n < 3 {
			return nil, fmt.Errorf("insufficient fill color fields: %d", n)
		}
	default:
		return nil, fmt.Errorf("invalid fill color, should be #RGB, #RGBA, #RRGGBB or #RRGGBBAA")
	}

	return c, nil
}

func save(img image.Image, imgType, outType, destDir, srcName string) (err error) {
	outType, unsupOnly := strings.CutPrefix(outType, "unsup:")
	if (unsupOnly && (imgType != "webp")) || (outType == "same") {
		outType = imgType
	}

	oldExt := filepath.Ext(srcName)
	destName := fmt.Sprintf("%s.%s", srcName[:len(srcName)-len(oldExt)], outType)

	outFile, err := os.CreateTemp(destDir, destName)
	if err != nil {
		return fmt.Errorf("could not create temporary destination %q: %w", destName, err)
	}
	canRename := false
	defer func() {
		if defErr := outFile.Sync(); defErr != nil {
			err = fmt.Errorf("could not flush temporary destination %q: %w", destName, defErr)
		}
		if defErr := outFile.Close(); defErr != nil {
			err = fmt.Errorf("could not close temporary destination %q: %w", destName, defErr)
		}

		if canRename {
			if defErr := os.Rename(outFile.Name(), filepath.Join(destDir, destName)); defErr != nil {
				err = fmt.Errorf("could not rename destination file %q: %w", destName, defErr)
			}
		}
	}()

	switch outType {
	case "gif":
		if err = gif.Encode(outFile, img, nil); err != nil {
			return fmt.Errorf("could not encode GIF destination %q: %w", destName, err)
		}
	case "jpeg":
		if err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100}); err != nil {
			return fmt.Errorf("could not encode JPEG destination %q: %w", destName, err)
		}
	case "png":
		enc := png.Encoder{
			CompressionLevel: png.BestCompression,
			BufferPool:       pngPool,
		}
		if err = enc.Encode(outFile, img); err != nil {
			return fmt.Errorf("could not encode PNG destination %q: %w", destName, err)
		}
	case "bmp":
		if err = bmp.Encode(outFile, img); err != nil {
			return fmt.Errorf("could not encode BMP destination %q: %w", destName, err)
		}
	case "tiff":
		if err = tiff.Encode(outFile, img, nil); err != nil {
			return fmt.Errorf("could not encode TIFF destination %q: %w", destName, err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outType)
	}

	canRename = true
	return err
}

type pngEncoderBufferPool struct {
	pool sync.Pool
}

// type pngEncoderBufferPool sync.Pool
func (p *pngEncoderBufferPool) Get() *png.EncoderBuffer {
	return p.pool.Get().(*png.EncoderBuffer)
}

func (p *pngEncoderBufferPool) Put(buf *png.EncoderBuffer) {
	p.pool.Put(buf)
}

var pngPool = &pngEncoderBufferPool{
	pool: sync.Pool{
		New: func() any {
			return &png.EncoderBuffer{}
		},
	},
}
