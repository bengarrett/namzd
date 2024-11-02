package ls

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/fastwalk"
)

type Config struct {
	Casesensitive bool
	Count         bool
	Directories   bool
	Follow        bool
	Quiet         bool
	Panic         bool
	NumWorkers    int
	Sort          fastwalk.SortMode
}

func (opt Config) Walks(w io.Writer, s string, roots ...string) error {
	count := 0
	var err error
	for _, root := range roots {
		if count, err = opt.Walk(w, count, s, root); err != nil {
			if !opt.Quiet {
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

func (opt Config) Walk(w io.Writer, count int, s, root string) (int, error) {
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
			if !opt.Quiet {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return nil
		}
		if match, err := opt.Match(s, d); !match {
			return err
		}
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

func (opt Config) Match(pattern string, d fs.DirEntry) (bool, error) {
	if d.IsDir() && !opt.Directories {
		return false, nil
	}
	name := d.Name()
	if !opt.Casesensitive {
		pattern = strings.ToLower(pattern)
		name = strings.ToLower(name)
	}
	return filepath.Match(pattern, name)
}
