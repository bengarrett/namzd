package cp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bengarrett/namzd/cp"
	"github.com/nalgeon/be"
)

func TestCheckDest(t *testing.T) {
	t.Parallel()
	t.Run("ValidDirectory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir() // Create a temporary directory
		err := cp.CheckDest(dir)
		be.Err(t, err, nil)
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		t.Parallel()
		dir := filepath.Join(os.TempDir(), "nonexistent_dir")
		err := cp.CheckDest(dir)
		be.True(t, err != nil)
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
		be.True(t, err != nil)
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

// Benchmarks

// BenchmarkCopy benchmarks file copying operation.
func BenchmarkCopy(b *testing.B) {
	src := filepath.Join("..", "testdata", "archive.tar")
	b.Run("SmallFile", func(b *testing.B) {
		for b.Loop() {
			dst := filepath.Join(b.TempDir(), "archive.tar")
			_ = cp.Copy(src, dst)
		}
	})
}

// BenchmarkCheckDest benchmarks destination directory validation.
func BenchmarkCheckDest(b *testing.B) {
	tmpDir := b.TempDir()
	b.ResetTimer()
	for b.Loop() {
		_ = cp.CheckDest(tmpDir)
	}
}
