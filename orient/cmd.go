package orient

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

type OpParams struct {
	Scan      string `help:"Source folder to scan" default:"."`
	Portrait  string `help:"Destination folder for portrait images" default:"portrait"`
	Landscape string `help:"Destination folder for landscape images" default:"landscape"`
}

type CLICmd struct {
	Cp struct {
		OpParams
	} `cmd:"" help:"Copy images to their respective folders"`
	Mv struct {
		OpParams
	} `cmd:"" help:"Move images to their respective folders"`
}

func (c *CLICmd) Validate(kctx *kong.Context) error {
	var conf *OpParams
	switch kctx.Selected().Name {
	case "cp":
		conf = &c.Cp.OpParams
	case "mv":
		conf = &c.Mv.OpParams
	}

	scanDir, err := filepath.Abs(conf.Scan)
	var info os.FileInfo
	if err == nil {
		if info, err = os.Stat(scanDir); err == nil && !info.IsDir() {
			err = fmt.Errorf("not a directory")
		}
	}
	if err != nil {
		return fmt.Errorf("invalid scan path %q: %w", conf.Scan, err)
	}
	conf.Scan = scanDir

	if !filepath.IsAbs(conf.Portrait) {
		conf.Portrait = filepath.Join(scanDir, conf.Portrait)
	}

	if !filepath.IsAbs(conf.Landscape) {
		conf.Landscape = filepath.Join(scanDir, conf.Landscape)
	}

	return nil
}

func (c *CLICmd) Run(subCmd string) error {
	var conf OpParams
	var fileOp func(string, string) error
	switch subCmd {
	case "cp":
		conf = c.Cp.OpParams
		fileOp = copyFile
	case "mv":
		conf = c.Mv.OpParams
		fileOp = moveFile
	}

	if err := os.MkdirAll(conf.Portrait, os.ModeDir); err != nil {
		return fmt.Errorf("unable to create portrait destination folder %q: %w", conf.Portrait, err)
	}

	if err := os.MkdirAll(conf.Landscape, os.ModeDir); err != nil {
		return fmt.Errorf("unable to create landscape destination folder %q: %w", conf.Landscape, err)
	}

	files, err := os.ReadDir(conf.Scan)
	if err != nil {
		return fmt.Errorf("unable to read folder %q: %w", conf.Scan, err)
	}

	var portraitCount, landscapeCount, errCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := filepath.Join(conf.Scan, file.Name())
		img, err := os.Open(name)
		if err != nil {
			slog.Error("could not open image", "file", name, "error", err)
			continue
		}

		imgConf, _, err := image.DecodeConfig(img)
		if err != nil {
			slog.Error("could not read image", "file", name, "error", err)
			continue
		}
		if err = img.Close(); err != nil {
			slog.Error("could not close image", "file", name, "error", err)
		}

		isPortrait := imgConf.Height > imgConf.Width
		var dest string
		if isPortrait {
			portraitCount++
			dest = filepath.Join(conf.Portrait, file.Name())
		} else {
			landscapeCount++
			dest = filepath.Join(conf.Landscape, file.Name())
		}

		if err = fileOp(name, dest); err != nil {
			errCount++
			slog.Error("could not operate image", "from", name, "to", dest, "error", err)
		}
	}

	slog.Info("stats", "portraits", portraitCount, "landscapes", landscapeCount, "errors", errCount, "total",
		portraitCount+landscapeCount)

	if errCount > 0 {
		return fmt.Errorf("error processing %d files", errCount)
	}
	return nil
}
