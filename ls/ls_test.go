package ls_test

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bengarrett/namzd/ls"
	"github.com/nalgeon/be"
)

func TestConfig_Copier(t *testing.T) {
	t.Parallel()
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	tests := []struct {
		name         string
		path         string
		opt          ls.Config
		wantContains []string
	}{
		{
			name: "Invalid path",
			path: "invalid_path",
			opt: ls.Config{
				Destination: t.TempDir(),
			},
			wantContains: []string{"no such file or directory"},
		},
		{
			name: "Invalid drestination",
			path: tdir,
			opt: ls.Config{
				Destination: "invalid_destionation",
			},
			wantContains: []string{"no such file or directory"},
		},
		{
			name: "Copy a directory",
			path: tdir,
			opt: ls.Config{
				Destination: t.TempDir(),
			},
			wantContains: []string{"is a directory"},
		},
		{
			name: "Copy a file",
			path: filepath.Join(tdir, "archive.tar"),
			opt: ls.Config{
				Destination: t.TempDir(),
			},
			wantContains: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			tt.opt.StdErrors = true      // Enable standard error output
			tt.opt.Copier(&buf, tt.path) // copier does not return anything
			for _, want := range tt.wantContains {
				got := buf.String()
				be.True(t, strings.Contains(got, want))
			}
		})
	}
}

func TestConfig(t *testing.T) { //nolint:funlen
	t.Parallel()
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	fmatches := []string{
		"file_1996", "file_1997.xyz", "file_2002.zyx",
		"archive.tar.xz", "archive.zip", "archive.tar",
	}
	tests := []struct {
		name         string
		pattern      string
		path         string
		opt          ls.Config
		wantContains []string
		wantErr      bool
	}{
		{
			name:    "Invalid directory",
			pattern: "*",
			path:    "/invalid_path",
			opt: ls.Config{
				StdErrors: true,
				Panic:     true,
			},
			wantContains: []string{""},
			wantErr:      true,
		},
		{
			name:         "A directory",
			pattern:      "*",
			path:         tdir,
			opt:          ls.Config{},
			wantContains: fmatches,
			wantErr:      false,
		},
		{
			name:    "Directory and archives",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive: true,
			},
			wantContains: []string{"file_1985 > ", "file_2012.txt > "},
			wantErr:      false,
		},
		{
			name:    "Directory and archives plus display modtime",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive:      true,
				LastModified: true,
			},
			wantContains: []string{
				"file_1985 (1985-01-01) > ",
				"file_2012.txt (2012-01-01) > ",
			},
			wantErr: false,
		},
		{
			name:    "Directory and archives plus display count, modtime",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive:      true,
				Count:        true,
				LastModified: true,
			},
			wantContains: []string{
				"file_1985 (1985-01-01) > " + filepath.Join(tdir, "archive.zip"),
				"file_2012.txt (2012-01-01) > " + filepath.Join(tdir, "archive.tar"),
			},
			wantErr: false,
		},
		{
			name:    "Directories files and archives plus display count modtime",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive:      true,
				Count:        true,
				Directory:    true,
				LastModified: true,
			},
			wantContains: []string{
				"file_1985 (1985-01-01) > " + filepath.Join(tdir, "archive.zip"),
				"file_2012.txt (2012-01-01) > " + filepath.Join(tdir, "archive.tar"),
			},
			wantErr: false,
		},
		{
			name:    "Oldest",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive: true,
				Count:   true,
				Oldest:  true,
			},
			wantContains: []string{
				"Oldest found match:",
				"file_1985 (1985-01-01) > " + filepath.Join(tdir, "archive.zip"),
			},
			wantErr: false,
		},
		{
			name:    "Newest",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Archive: true,
				Count:   true,
				Newest:  true,
			},
			wantContains: []string{
				"Newest found match:",
				"file_2002.zyx (2025-03-26) > " + filepath.Join(tdir, "file_2002.zyx"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			const resetCount = 0
			var buf bytes.Buffer
			_, err := tt.opt.Walk(&buf, resetCount, tt.pattern, tt.path)
			be.Equal(t, (err != nil), tt.wantErr)
			for _, want := range tt.wantContains {
				got := buf.String()
				be.True(t, strings.Contains(got, want))
			}
		})
	}
}

func TestConfig_Walk(t *testing.T) {
	t.Parallel()
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	tests := []struct {
		name      string
		pattern   string
		path      string
		opt       ls.Config
		wantFinds int
		wantErr   bool
	}{
		{
			name:      "Invalid path",
			pattern:   "*",
			path:      "invalid_path",
			opt:       ls.Config{},
			wantFinds: 0,
			wantErr:   true,
		},
		{
			name:      "Search within a directory",
			pattern:   "*",
			path:      tdir,
			opt:       ls.Config{},
			wantFinds: 6,
			wantErr:   false,
		},
		{
			name:    "Search within a directory and archives",
			pattern: "file_*",
			path:    tdir,
			opt: ls.Config{
				Archive: true,
			},
			wantFinds: 5, // 3 files in the directory and 2 in the archives
			wantErr:   false,
		},
		{
			name:    "Match files and directories",
			pattern: "*",
			path:    tdir,
			opt: ls.Config{
				Directory: true,
			},
			wantFinds: 8, // 7 files + 1 directory (testdata itself)
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			const resetCount = 0
			tt.opt.Count = true      // Enable counting otherwise 0 is always returned
			tt.opt.StdErrors = false // Disable standard error output
			count, err := tt.opt.Walk(io.Discard, resetCount, tt.pattern, tt.path)
			be.Equal(t, (err != nil), tt.wantErr)
			be.Equal(t, count, tt.wantFinds)
		})
	}
}

func TestDosEpoch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{
			name: "Before DOS epoch",
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: true,
		},
		{
			name: "Exactly at DOS epoch",
			time: time.Date(1980, 2, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "After DOS epoch",
			time: time.Date(1980, 2, 1, 0, 0, 1, 0, time.UTC),
			want: false,
		},
		{
			name: "Far after DOS epoch",
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, ls.DosEpoch(tt.time), tt.want)
		})
	}
}

func TestPrint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		count int
		path  string
		fd    ls.Find
		want  string
	}{
		{
			name:  "Print with count and mod time",
			count: 1,
			path:  "/path/to/file",
			fd: ls.Find{
				Name:    "file.txt",
				ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "1\tfile.txt (2023-10-01) > /path/to/file\n",
		},
		{
			name:  "Print without count and with mod time",
			count: 0,
			path:  "/path/to/file",
			fd: ls.Find{
				Name:    "file.txt",
				ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "file.txt (2023-10-01) > /path/to/file\n",
		},
		{
			name:  "Print with count and without mod time",
			count: 1,
			path:  "/path/to/file",
			fd: ls.Find{
				Name: "file.txt",
			},
			want: "1\tfile.txt > /path/to/file\n",
		},
		{
			name:  "Print without count and mod time",
			count: 0,
			path:  "/path/to/file",
			fd: ls.Find{
				Name: "file.txt",
			},
			want: "file.txt > /path/to/file\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			ls.Print(&buf, true, tt.count, tt.path, tt.fd)
			be.Equal(t, buf.String(), tt.want)
		})
	}
}

func TestMatch_Older(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    *ls.Match
		time time.Time
		want bool
	}{
		{
			name: "Older time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Newer time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.m.Older(tt.time), tt.want)
		})
	}
}

func TestMatch_Newer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    *ls.Match
		time time.Time
		want bool
	}{
		{
			name: "Newer time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Older time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.m.Newer(tt.time), tt.want)
		})
	}
}

func TestMatch_UpdateO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		m     ls.Match
		count int
		path  string
		fd    ls.Find
		want  ls.Match
	}{
		{
			name:  "Update older match",
			m:     ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  ls.Match{Count: 1, Path: "/path/to/file", Fd: ls.Find{Name: "file.txt", ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update newer match",
			m:     ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update with zero mod time",
			m:     ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Time{}},
			want:  ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
	}

	for i := range tests {
		tt := &tests[i] // Use pointer to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.m.UpdateO(tt.count, tt.path, tt.fd)
			be.Equal(t, tt.m.Count, tt.want.Count)
			be.Equal(t, tt.m.Path, tt.want.Path)
			be.Equal(t, tt.m.Fd, tt.want.Fd)
		})
	}
}

func TestMatch_UpdateN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		m     *ls.Match
		count int
		path  string
		fd    ls.Find
		want  *ls.Match
	}{
		{
			name:  "Update newer match",
			m:     &ls.Match{Fd: ls.Find{ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  &ls.Match{Count: 1, Path: "/path/to/file", Fd: ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update older match",
			m:     &ls.Match{Fd: ls.Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  &ls.Match{Fd: ls.Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update with zero mod time",
			m:     &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Time{}},
			want:  &ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
	}

	for i := range tests {
		tt := &tests[i] // Use pointer to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.m.UpdateN(tt.count, tt.path, tt.fd)
			be.Equal(t, tt.m.Count, tt.want.Count)
			be.Equal(t, tt.m.Path, tt.want.Path)
			be.Equal(t, tt.m.Fd, tt.want.Fd)
		})
	}
}

func TestConfig_Walks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  ls.Config
		pattern string
		roots   []string
		wantErr bool
	}{
		{
			name: "Single root, no error",
			config: ls.Config{
				StdErrors: false,
				Panic:     false,
			},
			pattern: "*.go",
			roots:   []string{"./"},
			wantErr: false,
		},
		{
			name: "Multiple roots, no error",
			config: ls.Config{
				StdErrors: false,
				Panic:     false,
			},
			pattern: "*.go",
			roots:   []string{"./", "../"},
			wantErr: false,
		},
		{
			name: "Single root, with error",
			config: ls.Config{
				StdErrors: false,
				Panic:     false,
			},
			pattern: "*.go",
			roots:   []string{"./nonexistent"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := tt.config.Walks(&buf, tt.pattern, tt.roots...)
			be.Equal(t, (err != nil), tt.wantErr)
		})
	}
}

func TestConfig_Archiver(t *testing.T) {
	t.Parallel()
	atar := filepath.Join("..", "testdata", "archive.tar")
	atxz := filepath.Join("..", "testdata", "archive.tar.xz")
	azip := filepath.Join("..", "testdata", "archive.zip")
	tests := []struct {
		name      string
		pattern   string
		path      string
		opt       ls.Config
		wantFinds int
		wantErr   bool
	}{
		{
			name:      "Invalid path",
			pattern:   "*",
			path:      "invalid_path",
			opt:       ls.Config{},
			wantFinds: 0,
			wantErr:   false,
		},
		{
			name:      "Search within an uncompress TAR archive",
			pattern:   "file_2012*",
			path:      atar,
			opt:       ls.Config{},
			wantFinds: 1,
			wantErr:   false,
		},
		{
			name:      "Search within a compress TAR.XZ archive",
			pattern:   "file_2012*",
			path:      atxz,
			opt:       ls.Config{},
			wantFinds: 0,
			wantErr:   false,
		},
		{
			name:      "Search within an compress ZIP archive",
			pattern:   "file_1985*",
			path:      azip,
			opt:       ls.Config{},
			wantFinds: 1,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.opt.Archive = true
			path, _ := filepath.Abs(tt.path)
			finds, err := tt.opt.Archiver(tt.pattern, path)
			be.Equal(t, (err != nil), tt.wantErr)
			be.Equal(t, len(finds), tt.wantFinds)
		})
	}
}

// Benchmarks

// BenchmarkConfig_Walk_CaseSensitive benchmarks directory walking with case-sensitive matching.
func BenchmarkConfig_Walk_CaseSensitive(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	opt := ls.Config{
		Casesensitive: true,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Walk(&buf, 0, "*.xyz", tdir)
	}
}

// BenchmarkConfig_Walk_CaseInsensitive benchmarks directory walking with case-insensitive matching.
func BenchmarkConfig_Walk_CaseInsensitive(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	opt := ls.Config{
		Casesensitive: false,
	}
	var buf bytes.Buffer
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Walk(&buf, 0, "*.xyz", tdir)
	}
}

// BenchmarkConfig_Walk_Wildcard benchmarks directory walking with wildcard patterns.
func BenchmarkConfig_Walk_Wildcard(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	opt := ls.Config{}
	var buf bytes.Buffer
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Walk(&buf, 0, "*", tdir)
	}
}

// BenchmarkConfig_Walk_LiteralMatch benchmarks directory walking with literal filename matching.
func BenchmarkConfig_Walk_LiteralMatch(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata"))
	opt := ls.Config{}
	var buf bytes.Buffer
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Walk(&buf, 0, "file_1996", tdir)
	}
}

// BenchmarkConfig_Archiver_Zip benchmarks archive detection for ZIP files.
func BenchmarkConfig_Archiver_Zip(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata", "archive.zip"))
	opt := ls.Config{
		Archive: true,
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Archiver("*", tdir)
	}
}

// BenchmarkConfig_Archiver_Tar benchmarks archive detection for TAR files.
func BenchmarkConfig_Archiver_Tar(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata", "archive.tar"))
	opt := ls.Config{
		Archive: true,
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Archiver("*", tdir)
	}
}

// BenchmarkConfig_Archiver_TarXZ benchmarks archive detection for TAR.XZ files.
func BenchmarkConfig_Archiver_TarXZ(b *testing.B) {
	tdir, _ := filepath.Abs(filepath.Join("..", "testdata", "archive.tar.xz"))
	opt := ls.Config{
		Archive: true,
	}
	b.ResetTimer()
	for b.Loop() {
		_, _ = opt.Archiver("*", tdir)
	}
}

// BenchmarkMatch_Older benchmarks the Older comparison method.
func BenchmarkMatch_Older(b *testing.B) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	m := &ls.Match{Fd: ls.Find{ModTime: t2}}
	b.ResetTimer()
	for b.Loop() {
		_ = m.Older(t1)
	}
}

// BenchmarkMatch_Newer benchmarks the Newer comparison method.
func BenchmarkMatch_Newer(b *testing.B) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	m := &ls.Match{Fd: ls.Find{ModTime: t2}}
	b.ResetTimer()
	for b.Loop() {
		_ = m.Newer(t1)
	}
}
