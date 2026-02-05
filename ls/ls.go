// Package ls provides file search and matching functionality across directories and archives.
package ls

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bengarrett/namzd/cp"
	"github.com/charlievieth/fastwalk"
)

// Config is the configuration for the ls command.
type Config struct {
	Archive       bool
	Casesensitive bool
	Count         bool
	Directory     bool
	Follow        bool
	LastModified  bool
	Oldest        bool
	Newest        bool
	Panic         bool
	StdErrors     bool
	Destination   string
	NumWorkers    int
	Sort          fastwalk.SortMode
}

// Walks the root directory paths to match filenames to the pattern and writes the results to the writer.
func (opt Config) Walks(w io.Writer, pattern string, roots ...string) error {
	count := 0
	var err error
	for _, root := range roots {
		if count, err = opt.Walk(w, count, pattern, root); err != nil {
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
			}
			if opt.Panic {
				os.Exit(1)
			}
			return fmt.Errorf("walks: %w", err)
		}
	}
	return nil
}

// Walk the root directory to match filenames to the pattern and writes the results to the out writer.
// The counted finds is returned or left at 0 if not counting.
func (opt Config) Walk(out io.Writer, count int, pattern, root string) (int, error) { //nolint:gocognit,cyclop,funlen
	conf := fastwalk.Config{
		Follow:     opt.Follow,
		Sort:       opt.Sort,
		NumWorkers: opt.NumWorkers,
	}
	oldest, newest := Match{}, Match{}
	var printMu sync.Mutex
	walkFn := func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			if opt.Panic {
				return fmt.Errorf("walk: %w", err)
			}
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return nil
		}
		finds, err := opt.Archiver(pattern, path)
		if err != nil {
			if opt.Panic {
				return fmt.Errorf("walk: %w", err)
			}
			if opt.StdErrors {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			}
			return nil
		}
		if len(finds) > 0 {
			for _, find := range finds {
				if opt.Count {
					count++
				}
				Print(out, opt.LastModified, count, path, find)
				if opt.Oldest {
					(&oldest).UpdateO(count, path, find)
				}
				if opt.Newest {
					(&newest).UpdateN(count, path, find)
				}
			}
			return nil
		}
		if match, err := opt.Match(pattern, dir.Name(), dir.IsDir()); !match {
			return err
		}
		opt.Copier(os.Stderr, path)
		opt.Update(dir, count, path, &oldest, &newest)
		if opt.Count {
			count++
		}
		// this is required to avoid a possible race condition when writing to the io.Writer
		printMu.Lock()
		defer printMu.Unlock()
		opt.Print(out, count, path)
		return err
	}
	if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
		return count, fmt.Errorf("fastwalk: %w", err)
	}
	opt.oldest(out, &oldest)
	opt.newest(out, &newest)
	return count, nil
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
	ok, err := filepath.Match(pattern, name)
	if err != nil {
		return false, fmt.Errorf("match: %w", err)
	}
	return ok, nil
}

// Copier copies the file to the destination directory path.
// The out writer is used to print the errors.
func (opt Config) Copier(out io.Writer, path string) {
	if opt.Destination == "" || opt.Archive {
		return
	}
	dest := filepath.Join(opt.Destination, filepath.Base(path))
	err := cp.Copy(path, dest)
	if err != nil && opt.StdErrors {
		fmt.Fprintf(out, "%s: %v\n", path, err)
	}
}

// Update the oldest and newest matches with the count, filename, path and file info.
// If the oldest and newest flags are not set or there is an error, the function exits.
func (opt Config) Update(dir fs.DirEntry, count int, path string, oldest, newest *Match) {
	if !opt.Oldest && !opt.Newest {
		return
	}
	if oldest == nil || newest == nil {
		return
	}
	info, err := dir.Info()
	if err != nil {
		return
	}
	if opt.Oldest {
		oldest.UpdateO(count, path, Find{
			Name:    dir.Name(),
			ModTime: info.ModTime(),
		})
	}
	if opt.Newest {
		newest.UpdateN(count, path, Find{
			Name:    dir.Name(),
			ModTime: info.ModTime(),
		})
	}
}

// Find is the matched filename and last modification time of the file.
type Find struct {
	Name    string
	ModTime time.Time
}

// Archiver reads the index of a tar or zip archive and returns the matched filenames to the pattern.
func (opt Config) Archiver(pattern, path string) ([]Find, error) {
	if !opt.Archive {
		return nil, nil
	}
	if ZipArchive(path) {
		return opt.Zips(pattern, path)
	}
	if TarArchive(path) {
		return opt.Tars(pattern, path)
	}
	return nil, nil
}

// Tars reads the index of the tar archive and returns the matched filenames to the pattern.
func (opt Config) Tars(pattern, path string) ([]Find, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("tars: %w", err)
	}
	defer file.Close()
	tr := tar.NewReader(file)
	finds := []Find{}
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tars: %w", err)
		}
		if match, _ := opt.Match(pattern, header.Name, header.FileInfo().IsDir()); !match {
			continue
		}
		finds = append(finds, Find{
			Name:    header.Name,
			ModTime: header.ModTime,
		})
	}
	return finds, nil
}

// Zips reads the index of the zip archive and returns the matched filenames to the pattern.
func (opt Config) Zips(pattern, path string) ([]Find, error) {
	zipper, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("zips: %w", err)
	}
	defer zipper.Close()
	finds := []Find{}
	for _, file := range zipper.File {
		info := file.FileInfo()
		if match, _ := opt.Match(pattern, file.Name, info.IsDir()); !match {
			continue
		}
		finds = append(finds, Find{
			Name:    file.Name,
			ModTime: file.Modified,
		})
	}
	return finds, nil
}

// TarArchive checks if the file is a tar archive.
// TarArchive checks if the file is a tar archive.
func TarArchive(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	const size, offset = 4, 257
	magic := make([]byte, size+offset)
	if _, err := file.Read(magic); err != nil {
		return false
	}
	// Check if the magic number matches the TAR file magic number
	if magic[offset+0] != 0x75 || magic[offset+1] != 0x73 ||
		magic[offset+2] != 0x74 || magic[offset+3] != 0x61 {
		return false
	}
	return true
}

// ZipArchive checks if the file is a zip archive.
func ZipArchive(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	const size = 4
	magic := make([]byte, size)
	if _, err := file.Read(magic); err != nil {
		return false
	}
	// Check if the magic number matches the ZIP file magic number
	if len(magic) < size || magic[0] != 0x50 || magic[1] != 0x4B ||
		magic[2] != 0x03 || magic[3] != 0x04 {
		return false
	}
	return true
}

// Match is the matched filename and path.
// It is used when the oldest or newest flags are set.
type Match struct {
	sync.Mutex

	Count int    // Count of matches.
	Path  string // Path to the matched file.
	Fd    Find   // Find of the matched file.
}

// Older compares the time is older than the match.
func (m *Match) Older(cmp time.Time) bool {
	if DosEpoch(cmp) {
		return false
	}
	if m.Fd.ModTime.IsZero() {
		return true
	}
	return cmp.Before(m.Fd.ModTime)
}

// Newer compares the time is newer than the match.
func (m *Match) Newer(cmp time.Time) bool {
	if DosEpoch(cmp) {
		return false
	}
	if m.Fd.ModTime.IsZero() {
		return true
	}
	return cmp.After(m.Fd.ModTime)
}

// UpdateO the match with the count, filename, path and file info if the modtime is older.
func (m *Match) UpdateO(count int, path string, find Find) {
	if find.ModTime.IsZero() || find.Name == "" {
		return
	}
	modTime := find.ModTime
	m.Lock()
	defer m.Unlock()
	// Inline the Older logic to avoid nested locking
	if DosEpoch(modTime) {
		return
	}
	if m.Fd.ModTime.IsZero() {
		// If current match has zero time, this file is older
		m.Count = count
		m.Path = path
		m.Fd = find
		return
	}
	if modTime.Before(m.Fd.ModTime) {
		m.Count = count
		m.Path = path
		m.Fd = find
	}
}

// UpdateN updates the match with the count, filename, path and file info if the modtime is newer.
func (m *Match) UpdateN(count int, path string, find Find) {
	if find.ModTime.IsZero() || find.Name == "" {
		return
	}
	modTime := find.ModTime
	m.Lock()
	defer m.Unlock()
	// Inline the Newer logic to avoid nested locking
	if DosEpoch(modTime) {
		return
	}
	if m.Fd.ModTime.IsZero() {
		// If current match has zero time, this file is newer
		m.Count = count
		m.Path = path
		m.Fd = find
		return
	}
	if modTime.After(m.Fd.ModTime) {
		m.Count = count
		m.Path = path
		m.Fd = find
	}
}

// Print the matched find to the writer.
func Print(out io.Writer, lastMod bool, count int, path string, find Find) {
	if out == nil {
		out = io.Discard
	}
	if count > 0 {
		fmt.Fprintf(out, "%d\t", count)
	}
	fmt.Fprint(out, find.Name)
	if lastMod && !find.ModTime.IsZero() {
		s := find.ModTime.Format("2006-01-02")
		fmt.Fprintf(out, " (%s)", s)
	}
	fmt.Fprint(out, " > "+path+"\n")
}

// Print the matched path to the out writer.
func (opt Config) Print(out io.Writer, count int, path string) {
	if out == nil {
		out = io.Discard
	}
	if count > 0 {
		fmt.Fprintf(out, "%d\t", count)
	}
	if opt.LastModified {
		fi, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(out)
			return
		}
		s := fi.ModTime().Format("2006-01-02")
		fmt.Fprintf(out, "(%s) ", s)
	}
	fmt.Fprint(out, path+"\n")
}

func (opt Config) oldest(w io.Writer, oldest *Match) {
	if !opt.Oldest || oldest.Fd.ModTime.IsZero() || oldest.Fd.Name == "" {
		return
	}
	fmt.Fprintln(w, "Oldest found match:")
	Print(w, true, oldest.Count, oldest.Path, oldest.Fd)
}

func (opt Config) newest(w io.Writer, newest *Match) {
	if !opt.Newest || newest.Fd.ModTime.IsZero() || newest.Fd.Name == "" {
		return
	}
	fmt.Fprintln(w, "Newest found match:")
	Print(w, true, newest.Count, newest.Path, newest.Fd)
}

// DosEpoch checks if the time is before the MS-DOS epoch.
//
// However due to false positives created by systems that lacked a real-time clock,
// it treats Epoch as the 1 February 1980, and not 1 January, 1980.
func DosEpoch(t time.Time) bool {
	return t.UTC().Before(time.Date(1980, time.February, 1, 0, 0, 0, 0, time.UTC))
}
