package cp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CheckDest checks if the destination directory exists and is writable.
func CheckDest(dir string) error {
	if dir == "" {
		return nil
	}
	st, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("destination directory does not exist: %s", dir)
	}
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("destination is not a directory: %s", dir)
	}
	testFile := filepath.Join(dir, "namezed_test")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("destination directory is not writable: %s", err)
	}
	file.Close()
	defer os.Remove(testFile)
	return nil
}

// Copy a file from source to destination.
func Copy(source, destination string) error {
	st, err := os.Stat(source)
	if err != nil {
		return err
	}
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	if err := os.Chmod(destination, st.Mode()); err != nil {
		return err
	}

	return dst.Sync()
}
