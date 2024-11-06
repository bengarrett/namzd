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
		return fmt.Errorf("destination directory is not writable: %s", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
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
