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
			return err
		}
	}
	return nil
}

// Walk the root directory to match filenames to the pattern and writes the results to the writer.
// The counted finds is returned or left at 0 if not counting.
func (opt Config) Walk(w io.Writer, count int, pattern, root string) (int, error) {
	conf := fastwalk.Config{
		Follow:     opt.Follow,
		Sort:       opt.Sort,
		NumWorkers: opt.NumWorkers,
	}
	oldest, newest := Match{}, Match{}
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
			for _, fd := range finds {
				if opt.Count {
					count++
				}
				Print(w, opt.LastModified, count, path, fd)
				if opt.Oldest {
					oldest.UpdateO(count, path, fd)
				}
				if opt.Newest {
					newest.UpdateN(count, path, fd)
				}
			}
			return nil
		}
		if match, err := opt.Match(pattern, d.Name(), d.IsDir()); !match {
			return err
		}
		opt.Copier(os.Stderr, path)
		opt.Update(d, count, path, &oldest, &newest)
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
	if opt.Oldest && !oldest.Fd.ModTime.IsZero() && count > 1 {
		fmt.Fprintln(w, "Oldest found match:")
		Print(w, true, oldest.Count, oldest.Path, oldest.Fd)
	}
	if opt.Newest && !newest.Fd.ModTime.IsZero() && count > 1 {
		fmt.Fprintln(w, "Newest found match:")
		Print(w, true, newest.Count, newest.Path, newest.Fd)
	}
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
	return filepath.Match(pattern, name)
}

// Copier copies the file to the destination directory.
func (opt Config) Copier(w io.Writer, path string) {
	if opt.Destination == "" || opt.Archive {
		return
	}
	defer func() {
		dest := filepath.Join(opt.Destination, filepath.Base(path))
		if err := cp.Copy(path, dest); err != nil {
			if opt.StdErrors {
				fmt.Fprintf(w, "%s: %v\n", path, err)
			}
			return
		}
	}()
}

// Update the oldest and newest matches with the count, filename, path and file info.
// If the oldest and newest flags are not set or there is an error, the function exits.
func (opt Config) Update(d fs.DirEntry, count int, path string, oldest, newest *Match) {
	if !opt.Oldest && !opt.Newest {
		return
	}
	if oldest == nil || newest == nil {
		return
	}
	info, err := d.Info()
	if err != nil {
		return
	}
	if opt.Oldest {
		oldest.UpdateO(count, path, Find{
			Name:    d.Name(),
			ModTime: info.ModTime(),
		})
	}
	if opt.Newest {
		newest.UpdateN(count, path, Find{
			Name:    d.Name(),
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
		return nil, err
	}
	defer file.Close()
	tr := tar.NewReader(file)
	finds := []Find{}
	for {
		th, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if match, _ := opt.Match(pattern, th.Name, th.FileInfo().IsDir()); !match {
			continue
		}
		finds = append(finds, Find{
			Name:    th.Name,
			ModTime: th.ModTime,
		})
	}
	return finds, nil
}

// Zips reads the index of the zip archive and returns the matched filenames to the pattern.
func (opt Config) Zips(pattern, path string) ([]Find, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	finds := []Find{}
	for _, f := range r.File {
		fi := f.FileInfo()
		if match, _ := opt.Match(pattern, f.Name, fi.IsDir()); !match {
			continue
		}
		finds = append(finds, Find{
			Name:    f.Name,
			ModTime: f.Modified,
		})
	}
	return finds, nil
}

// TarArchive checks if the file is a tar archive.
func TarArchive(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	const offset = 257
	magic := make([]byte, 4+offset)
	if _, err := file.Read(magic); err != nil {
		return false
	}
	// Check if the magic number matches the TAR file magic number
	if magic[offset+0] != 0x75 || magic[offset+1] != 0x73 ||
		magic[offset+2] != 0x74 || magic[offset+3] != 0x61 {
		return false
	}
	file.Close()
	return true
}

// ZipArchive checks if the file is a zip archive.
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

// Match is the matched filename and path.
// It is used when the oldest or newest flags are set.
type Match struct {
	Count int    // Count of matches.
	Path  string // Path to the matched file.
	Fd    Find   // Find of the matched file.
}

// Older checks if the time is older than the match.
func (m Match) Older(t time.Time) bool {
	if DosEpoch(t) {
		return false
	}
	if m.Fd.ModTime.IsZero() {
		return true
	}
	return t.Before(m.Fd.ModTime)
}

// Newer checks if the time is newer than the match.
func (m Match) Newer(t time.Time) bool {
	if DosEpoch(t) {
		return false
	}
	if m.Fd.ModTime.IsZero() {
		return true
	}
	return t.After(m.Fd.ModTime)
}

// UpdateO the match with the count, filename, path and file info if the modtime is older.
func (m *Match) UpdateO(c int, path string, fd Find) {

	if fd.ModTime.IsZero() || fd.Name == "" {
		return
	}
	t := fd.ModTime
	if !m.Older(t) {
		return
	}
	m.Count = c
	m.Path = path
	m.Fd = fd
}

// Update the match with the count, filename, path and file info if the modtime is newer.
func (m *Match) UpdateN(c int, path string, fd Find) {
	if fd.ModTime.IsZero() || fd.Name == "" {
		return
	}
	t := fd.ModTime
	if !m.Newer(t) {
		return
	}
	m.Count = c
	m.Path = path
	m.Fd = fd
}

// Print the match to the writer.
func Print(w io.Writer, lastMod bool, count int, path string, fd Find) {
	if w == nil {
		w = io.Discard
	}
	if count > 0 {
		fmt.Fprintf(w, "%d\t", count)
	}
	fmt.Fprintf(w, "%s", fd.Name)
	if lastMod && !fd.ModTime.IsZero() {
		s := fd.ModTime.Format("2006-01-02")
		fmt.Fprintf(w, " (%s)", s)
	}
	fmt.Fprintf(w, " > %s\n", path)
}

// DosEpoch checks if the time is before the MS-DOS epoch.
//
// However due to false positives created by systems that lacked a real-time clock,
// it treats Epoch as the 1 February 1980, and not 1 January, 1980.
func DosEpoch(t time.Time) bool {
	return t.UTC().Before(time.Date(1980, 2, 0, 0, 0, 0, 0, time.UTC))
}
