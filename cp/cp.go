// Package cp provides file copying utilities for matched files.
package cp

import (
	"fmt"
	"io"
	"os"
)

// CheckDest checks if the destination directory exists and is writable.
func CheckDest(dir string) error {
	tmp, err := os.CreateTemp(dir, "namezed_test")
	if err != nil {
		return fmt.Errorf("destination directory is not writable: %w", err)
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()
	return nil
}

// Copy a file from source to destination.
func Copy(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	src, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	defer dst.Close()

	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		os.Remove(destination)
		return fmt.Errorf("copy: %w", err)
	}

	if err := os.Chmod(destination, info.Mode()); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
