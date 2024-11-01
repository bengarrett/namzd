package ls

import (
	"fmt"
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
	NumWorkers    int
	Sort          fastwalk.SortMode
}

func (opt Config) Walks(s string, roots ...string) error {
	count := 0
	var err error
	for _, root := range roots {
		if count, err = opt.Walk(count, s, root); err != nil {
			return err
		}
	}
	return nil
}

func (opt Config) Walk(count int, s, root string) (int, error) {
	conf := fastwalk.Config{
		Follow:     opt.Follow,
		Sort:       opt.Sort,
		NumWorkers: opt.NumWorkers,
	}
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			return nil // returning the error stops iteration
		}
		if s == "" {
			return nil
		}
		pattern := s
		name := d.Name()
		if d.IsDir() && !opt.Directories {
			return nil
		}
		if !opt.Casesensitive {
			pattern = strings.ToLower(s)
			name = strings.ToLower(name)
		}
		if match, err := filepath.Match(pattern, name); !match {
			return err
		}
		if opt.Count {
			count++
			_, err = fmt.Printf("%d\t%s\n", count, path)
			return err
		}
		// todo: replace with stdout
		_, err = fmt.Println(path)
		return err
	}
	if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
		os.Exit(1) // todo: option to exit or continue
	}
	return count, nil
}
