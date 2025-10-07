package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

type runConfig struct {
	cmd          string
	scanDir      string
	portraitDir  string
	landscapeDir string
}

func getConfig() (runConfig, error) {
	n := len(os.Args)
	if n < 2 {
		return runConfig{}, fmt.Errorf("not enough parameters")
	}

	conf := runConfig{
		cmd: os.Args[1],
	}

	switch conf.cmd {
	case "orientcp":
		var err error

		scanDir := "."
		if n > 2 {
			scanDir = os.Args[2]
		}
		conf.scanDir, err = filepath.Abs(scanDir)
		if err != nil {
			return runConfig{}, fmt.Errorf("invalid scan path %q: %w", os.Args[2], err)
		}

		portraitDir := filepath.Join(conf.scanDir, "portrait")
		if n > 3 {
			portraitDir = os.Args[3]
		}
		conf.portraitDir, err = filepath.Abs(portraitDir)
		if err != nil {
			return runConfig{}, fmt.Errorf("invalid portrait path %q: %w", os.Args[2], err)
		}

		landscapeDir := filepath.Join(conf.scanDir, "landscape")
		if n > 4 {
			landscapeDir = os.Args[4]
		}
		conf.landscapeDir, err = filepath.Abs(landscapeDir)
		if err != nil {
			return runConfig{}, fmt.Errorf("invalid landscape path %q: %w", os.Args[2], err)
		}
	default:
		return runConfig{}, fmt.Errorf("unsupported operation")
	}

	return conf, nil
}

func help() {
	slog.Info(fmt.Sprintf("%s orientcp [portrait_dir] [landscape_dir]\n", os.Args[0]))
}

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}
	conf, err := getConfig()
	if err != nil {
		slog.Error("invalid configuration", "err", err)
		return
	}

	slog.Info("running", "config", conf)

	files, err := os.ReadDir(conf.scanDir)
	if err != nil {
		slog.Error("unable to read folder", "dir", conf.scanDir, "error", err)
		return
	}

	if err := os.MkdirAll(conf.portraitDir, os.ModeDir); err != nil {
		slog.Error("unable to create portrait destination folder", "dir", conf.portraitDir, "error", err)
		return
	}

	if err := os.MkdirAll(conf.landscapeDir, os.ModeDir); err != nil {
		slog.Error("unable to create landscape destination folder", "dir", conf.landscapeDir, "error", err)
		return
	}

	var portraitCount, landscapeCount, errCount int
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := filepath.Join(conf.scanDir, file.Name())
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

		isLandscape := imgConf.Width >= imgConf.Height
		var dest string
		if isLandscape {
			landscapeCount++
			dest = filepath.Join(conf.landscapeDir, file.Name())
		} else {
			portraitCount++
			dest = filepath.Join(conf.portraitDir, file.Name())
		}

		if err = copyFile(name, dest); err != nil {
			errCount++
			slog.Error("could not copy image", "from", name, "to", dest, "error", err)
			continue
		}
	}

	slog.Info("stats", "portraits", portraitCount, "landscapes", landscapeCount, "errors", errCount, "total",
		portraitCount+landscapeCount)
}

func copyFile(src, dest string) error {
	slog.Info("copying", "from", src, "to", dest)

	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat source file %q: %w", src, err)
	}
	if !srcFileInfo.Mode().IsRegular() {
		return fmt.Errorf("cannot copy non-regular file %q: %s", srcFileInfo.Name(), srcFileInfo.Mode().String())
	}
	destFileInfo, err := os.Stat(dest)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cannot stat destination file %q: %w", dest, err)
		}
	} else {
		return fmt.Errorf("destination file already exists: %q", destFileInfo.Name())
	}

	inFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("could not open source file %q: %w", src, err)
	}
	defer func() {
		if close_err := inFile.Close(); close_err != nil {
			slog.Error("could not close source file", "name", src, "error", close_err)
		}
	}()

	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("could not open destination file %q: %w", dest, err)
	}
	defer func() {
		if close_err := outFile.Close(); close_err != nil {
			slog.Error("could not close destination file", "name", dest, "error", close_err)
		}
	}()

	if _, err = io.Copy(outFile, inFile); err != nil {
		return fmt.Errorf("could not copy from %q to %q: %w", src, dest, err)
	}

	err = outFile.Sync()
	if err != nil {
		return fmt.Errorf("could not flush destination file %q: %w", dest, err)
	}
	return nil
}
