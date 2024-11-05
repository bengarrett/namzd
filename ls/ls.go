package ls

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bengarrett/namezed/cp"
	"github.com/charlievieth/fastwalk"
)

type Config struct {
	Archive       bool
	Casesensitive bool
	Count         bool
	Directory     bool
	Follow        bool
	Panic         bool
	StdErrors     bool
	Destination   string
	NumWorkers    int
	Sort          fastwalk.SortMode
}

func (opt Config) Walks(w io.Writer, pattern string, roots ...string) error {
	count := 0
	var err error
	for _, root := range roots {
		if count, err = opt.Walk(w, count, pattern, root); err != nil {
			if opt.StdErrors {
				// hello
				fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
			}
			if opt.Panic {
				os.Exit(1)
			}
			return err
		}
	}
	return nil
}

func (opt Config) Walk(w io.Writer, count int, pattern, root string) (int, error) {
	conf := fastwalk.Config{
		Follow:     opt.Follow,
		Sort:       opt.Sort,
		NumWorkers: opt.NumWorkers,
	}
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if opt.Panic {
				return err
			}
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return nil
		}
		finds, err := opt.Archiver(pattern, path)
		if err != nil {
			if opt.Panic {
				return err
			}
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return nil
		}
		if len(finds) > 0 {
			for _, f := range finds {
				if opt.Count {
					count++
					fmt.Fprintf(w, "%d\t%s > %s\n", count, f, path)
					continue
				}
				fmt.Fprintf(w, "%s > %s", path, f)
			}
			return nil
		}
		if match, err := opt.Match(pattern, d.Name(), d.IsDir()); !match {
			return err
		}
		opt.Copier(path)
		if opt.Count {
			count++
			_, err = fmt.Fprintf(w, "%d\t%s\n", count, path)
			return err
		}
		_, err = fmt.Fprintln(w, path)
		return err
	}
	if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
		return count, err
	}
	return count, nil
}

// Copier copies the file to the destination directory.
func (opt Config) Copier(path string) {
	if opt.Destination == "" {
		return
	}
	defer func() {
		dest := filepath.Join(opt.Destination, filepath.Base(path))
		if err := cp.Copy(path, dest); err != nil {
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return
		}
	}()
}

func (opt Config) Archiver(pattern, path string) ([]string, error) {
	if !opt.Archive {
		return nil, nil
	}
	if !ZipArchive(path) {
		return nil, nil
	}
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	finds := []string{}
	for _, f := range r.File {
		fi := f.FileInfo()
		if match, _ := opt.Match(pattern, f.Name, fi.IsDir()); !match {
			continue
		}
		finds = append(finds, f.Name)
	}
	return finds, nil
}

func ZipArchive(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	magic := make([]byte, 4)
	if _, err := file.Read(magic); err != nil {
		return false
	}
	// Check if the magic number matches the ZIP file magic number
	if magic[0] != 0x50 || magic[1] != 0x4B ||
		magic[2] != 0x03 || magic[3] != 0x04 {
		return false
	}
	file.Close()
	return true
}

// Match checks if the file or directory matches the glob pattern or name.
func (opt Config) Match(pattern, filename string, isDir bool) (bool, error) {
	if isDir && !opt.Directory {
		return false, nil
	}
	name := filename
	if !opt.Casesensitive {
		pattern = strings.ToLower(pattern)
		name = strings.ToLower(name)
	}
	return filepath.Match(pattern, name)
}
