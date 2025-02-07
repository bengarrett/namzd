package ls

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestDosEpoch(t *testing.T) {
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
			if got := DosEpoch(tt.time); got != tt.want {
				t.Errorf("DosEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name  string
		count int
		path  string
		fd    Find
		want  string
	}{
		{
			name:  "Print with count and mod time",
			count: 1,
			path:  "/path/to/file",
			fd: Find{
				Name:    "file.txt",
				ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "1\tfile.txt (2023-10-01) > /path/to/file\n",
		},
		{
			name:  "Print without count and with mod time",
			count: 0,
			path:  "/path/to/file",
			fd: Find{
				Name:    "file.txt",
				ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			},
			want: "file.txt (2023-10-01) > /path/to/file\n",
		},
		{
			name:  "Print with count and without mod time",
			count: 1,
			path:  "/path/to/file",
			fd: Find{
				Name: "file.txt",
			},
			want: "1\tfile.txt > /path/to/file\n",
		},
		{
			name:  "Print without count and mod time",
			count: 0,
			path:  "/path/to/file",
			fd: Find{
				Name: "file.txt",
			},
			want: "file.txt > /path/to/file\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Print(&buf, tt.count, tt.path, tt.fd)
			if got := buf.String(); got != tt.want {
				t.Errorf("Print() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_Older(t *testing.T) {
	tests := []struct {
		name string
		m    Match
		time time.Time
		want bool
	}{
		{
			name: "Older time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Newer time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    Match{Fd: Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Older(tt.time); got != tt.want {
				t.Errorf("Match.Older() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_Newer(t *testing.T) {
	tests := []struct {
		name string
		m    Match
		time time.Time
		want bool
	}{
		{
			name: "Newer time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "Older time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "Zero mod time",
			m:    Match{Fd: Find{ModTime: time.Time{}}},
			time: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "DOS epoch time",
			m:    Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			time: time.Date(1979, 12, 31, 23, 59, 59, 0, time.UTC),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Newer(tt.time); got != tt.want {
				t.Errorf("Match.Newer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_UpdateO(t *testing.T) {
	tests := []struct {
		name  string
		m     Match
		count int
		path  string
		fd    Find
		want  Match
	}{
		{
			name:  "Update older match",
			m:     Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  Match{Count: 1, Path: "/path/to/file", Fd: Find{Name: "file.txt", ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update newer match",
			m:     Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update with zero mod time",
			m:     Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Time{}},
			want:  Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.UpdateO(tt.count, tt.path, tt.fd)
			if tt.m != tt.want {
				t.Errorf("Match.UpdateO() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestMatch_UpdateN(t *testing.T) {
	tests := []struct {
		name  string
		m     Match
		count int
		path  string
		fd    Find
		want  Match
	}{
		{
			name:  "Update newer match",
			m:     Match{Fd: Find{ModTime: time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  Match{Count: 1, Path: "/path/to/file", Fd: Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update older match",
			m:     Match{Fd: Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)},
			want:  Match{Fd: Find{ModTime: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
		{
			name:  "Do not update with zero mod time",
			m:     Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
			count: 1,
			path:  "/path/to/file",
			fd:    Find{Name: "file.txt", ModTime: time.Time{}},
			want:  Match{Fd: Find{ModTime: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.UpdateN(tt.count, tt.path, tt.fd)
			if tt.m != tt.want {
				t.Errorf("Match.UpdateN() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func TestTarArchive(t *testing.T) {
	tests := []struct {
		name     string
		fileData []byte
		want     bool
	}{
		{
			name:     "Valid tar archive",
			fileData: []byte{0x75, 0x73, 0x74, 0x61},
			want:     true,
		},
		{
			name:     "Invalid tar archive",
			fileData: []byte{0x00, 0x00, 0x00, 0x00},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "testtar")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(tt.fileData); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			if got := TarArchive(tmpfile.Name()); got != tt.want {
				t.Errorf("TarArchive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZipArchive(t *testing.T) {
	tests := []struct {
		name     string
		fileData []byte
		want     bool
	}{
		{
			name:     "Valid zip archive",
			fileData: []byte{0x50, 0x4B, 0x03, 0x04},
			want:     true,
		},
		{
			name:     "Invalid zip archive",
			fileData: []byte{0x00, 0x00, 0x00, 0x00},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "testzip")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(tt.fileData); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			if got := ZipArchive(tmpfile.Name()); got != tt.want {
				t.Errorf("ZipArchive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Walks(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		pattern string
		roots   []string
		wantErr bool
	}{
		{
			name: "Single root, no error",
			config: Config{
				StdErrors: false,
				Panic:     false,
			},
			pattern: "*.go",
			roots:   []string{"./"},
			wantErr: false,
		},
		{
			name: "Multiple roots, no error",
			config: Config{
				StdErrors: false,
				Panic:     false,
			},
			pattern: "*.go",
			roots:   []string{"./", "../"},
			wantErr: false,
		},
		{
			name: "Single root, with error",
			config: Config{
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
			var buf bytes.Buffer
			err := tt.config.Walks(&buf, tt.pattern, tt.roots...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Walks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
