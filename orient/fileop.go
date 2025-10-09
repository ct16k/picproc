package orient

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
)

func copyFile(src, dest string) error {
	slog.Info("copying", "from", src, "to", dest)

	if err := checkFile(src, dest); err != nil {
		return err
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

func moveFile(src, dest string) error {
	slog.Info("moving", "from", src, "to", dest)

	if err := checkFile(src, dest); err != nil {
		return err
	}

	return os.Rename(src, dest)
}

func checkFile(src, dest string) error {
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

	return nil
}
