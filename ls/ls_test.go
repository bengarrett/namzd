package ls_test

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bengarrett/namzd/ls"
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
				if got := buf.String(); !strings.Contains(got, want) {
					t.Errorf("Config.Walk() = %v, want %v", got, want)
				}
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
				"5	file_1985 (1985-01-01) > ",
				"6	file_2012.txt (2012-01-01) > ",
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
				"6	file_1985 (1985-01-01) > ",
				"7	file_2012.txt (2012-01-01) > ",
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
				"5	file_1985 (1985-01-01) > ",
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
				"3	archive.tar.xz (2025-02-08) >",
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Walk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, want := range tt.wantContains {
				if got := buf.String(); !strings.Contains(got, want) {
					t.Errorf("Config.Walk() = %v, want %v", got, want)
				}
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
			wantFinds: 8,
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Walk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if count != tt.wantFinds {
				t.Errorf("Config.Walk() = %v, want %v", count, tt.wantFinds)
			}
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
			if got := ls.DosEpoch(tt.time); got != tt.want {
				t.Errorf("DosEpoch() = %v, want %v", got, tt.want)
			}
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
			if got := buf.String(); got != tt.want {
				t.Errorf("Print() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_Older(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    ls.Match
		time time.Time
		want bool
	}{
		{
			name: "Older time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Newer time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.m.Older(tt.time); got != tt.want {
				t.Errorf("Match.Older() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_Newer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    ls.Match
		time time.Time
		want bool
	}{
		{
			name: "Newer time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Older time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    ls.Match{Fd: ls.Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.m.Newer(tt.time); got != tt.want {
				t.Errorf("Match.Newer() = %v, want %v", got, tt.want)
			}
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.m.UpdateO(tt.count, tt.path, tt.fd)
			if tt.m != tt.want {
				t.Errorf("Match.UpdateO() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMatch_UpdateN(t *testing.T) {
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
			name:  "Update newer match",
			m:     ls.Match{Fd: ls.Find{ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  ls.Match{Count: 1, Path: "/path/to/file", Fd: ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update older match",
			m:     ls.Match{Fd: ls.Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    ls.Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  ls.Match{Fd: ls.Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.m.UpdateN(tt.count, tt.path, tt.fd)
			if tt.m != tt.want {
				t.Errorf("Match.UpdateN() = %v, want %v", tt.m, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Walks() error = %v, wantErr %v", err, tt.wantErr)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Archiver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(finds) != tt.wantFinds {
				t.Errorf("Config.Archiver() = %v, want %v", len(finds), tt.wantFinds)
			}
		})
	}
}
