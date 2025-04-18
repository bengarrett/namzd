package cp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bengarrett/namzd/cp"
)

func TestCheckDest(t *testing.T) {
	t.Parallel()
	t.Run("ValidDirectory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir() // Create a temporary directory
		err := cp.CheckDest(dir)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(os.TempDir(), "nonexistent_dir")
		err := cp.CheckDest(dir)
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	})

	t.Run("NonWritableDirectory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		err := os.Chmod(dir, 0o555) // Make the directory read-only
		if err != nil {
			t.Fatalf("failed to set directory permissions: %v", err)
		}
		defer func() {
			_ = os.Chmod(dir, 0o755) // Restore permissions after the test
		}()

		err = cp.CheckDest(dir)
		if err == nil {
			t.Errorf("expected an error, got nil")
		}
	})
}

func TestCopy(t *testing.T) {
	t.Parallel()
	// Walk testdata directory
	testdata := filepath.Join("..", "testdata")
	err := filepath.Walk(testdata, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("walk error at %q: %v", path, err)
			return err
		}
		if info.IsDir() {
			t.Logf("Found: %s (dir: %v)", path, info.IsDir())
			return nil
		}
		t.Run("CopyFile", func(t *testing.T) {
			t.Parallel()
			dest := filepath.Join(t.TempDir(), info.Name())
			err := cp.Copy(path, dest)
			if err != nil {
				t.Errorf("failed to copy %s to %s: %v", path, dest, err)
			}
			t.Logf("Copied: %s to %s", path, dest)
		})
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk testdata directory: %v", err)
	}
}
