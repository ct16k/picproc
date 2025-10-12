package orient

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"picproc/parallel"

	"github.com/alecthomas/kong"
)

type OpParams struct {
	Scan      string `help:"Source folder to scan" default:"."`
	Portrait  string `help:"Destination folder for portrait images. Relative to scan dir if not absolute." default:"portrait"`
	Landscape string `help:"Destination folder for landscape images. Relative to scan dir if not absolute." default:"landscape"`
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
	if conf.Portrait == conf.Scan {
		return fmt.Errorf("source folder and portrait destination are the same")
	}

	if !filepath.IsAbs(conf.Landscape) {
		conf.Landscape = filepath.Join(scanDir, conf.Landscape)
	}
	switch conf.Landscape {
	case conf.Scan:
		return fmt.Errorf("source folder and landscape destination are the same")
	case conf.Portrait:
		return fmt.Errorf("portrait and landscape destinations are the same")
	}

	return nil
}

func (c *CLICmd) Run(subCmd string, worker parallel.WorkerFunc, wait parallel.WaitFunc) error {
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

	var portraitCount, landscapeCount, errCount atomic.Uint64
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		worker(func(fileName string) func() {
			return func() {
				filePath := filepath.Join(conf.Scan, fileName)
				imgFile, err := os.Open(filePath)
				if err != nil {
					errCount.Add(1)
					slog.Error("could not open image", "file", filePath, "error", err)
					return
				}

				imgConf, _, err := image.DecodeConfig(imgFile)
				if err != nil {
					errCount.Add(1)
					slog.Error("could not read image", "file", filePath, "error", err)
					return
				}
				if err = imgFile.Close(); err != nil {
					errCount.Add(1)
					slog.Error("could not close image", "file", filePath, "error", err)
					return
				}

				isPortrait := imgConf.Height > imgConf.Width
				var dest string
				var count *atomic.Uint64
				if isPortrait {
					count = &portraitCount
					dest = filepath.Join(conf.Portrait, fileName)
				} else {
					count = &landscapeCount
					dest = filepath.Join(conf.Landscape, fileName)
				}

				if err = fileOp(filePath, dest); err != nil {
					errCount.Add(1)
					slog.Error("could not operate image", "from", filePath, "to", dest, "error", err)
				} else {
					(*count).Add(1)
				}
			}
		}(file.Name()))
	}

	wait(true)

	portraits := portraitCount.Load()
	landscapes := landscapeCount.Load()
	errors := errCount.Load()
	slog.Info("stats", "portraits", portraits, "landscapes", landscapes, "errors", errors,
		"total", portraits+landscapes)

	if errors > 0 {
		return fmt.Errorf("error processing %d files", errors)
	}
	return nil
}
