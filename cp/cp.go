// Package cp provides file copying utilities for matched files.
package cp

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// CheckDest checks if the destination directory exists and is writable.
func CheckDest(dir string) (err error) {
	tmp, err := os.CreateTemp(dir, "namezed_test")
	if err != nil {
		return fmt.Errorf("destination directory is not writable: %w", err)
	}
	defer func() {
		if cErr := tmp.Close(); cErr != nil {
			err = errors.Join(err,
				fmt.Errorf("cannot close temporary file: %w", cErr))
		}
		if rErr := os.Remove(tmp.Name()); rErr != nil {
			err = errors.Join(err, fmt.Errorf("cannot remove temorary file: %w", rErr))
		}
	}()
	return nil
}

// Copy a file from source to destination.
func Copy(source, destination string) error {
	const format = "copy: %w"
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	src, err := os.Open(source)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	defer dst.Close()

	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		os.Remove(destination)
		return fmt.Errorf(format, err)
	}

	if err := os.Chmod(destination, info.Mode()); err != nil {
		return fmt.Errorf(format, err)
	}
	return nil
}
