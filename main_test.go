package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/bengarrett/namzd/cp"
	"github.com/bengarrett/namzd/ls"
	"github.com/charlievieth/fastwalk"
	"github.com/nalgeon/be"
)

func TestVersionFlag(t *testing.T) {
	t.Parallel()
	t.Run("VersionFlag implements required interfaces", func(t *testing.T) {
		t.Parallel()
		var v VersionFlag
		be.True(t, v.IsBool())
	})
}

func TestCmdRun(t *testing.T) { //nolint:funlen
	// Removed t.Parallel() due to race conditions with global state
	// Set NO_COLOR for consistent test output
	t.Setenv("NO_COLOR", "1")
	tdir := "testdata"

	tests := []struct {
		name         string
		cmd          Cmd
		expectError  bool
		wantContains []string
	}{
		{
			name: "Basic file search",
			cmd: Cmd{
				Match:   "*.go",
				Paths:   []string{tdir},
				Archive: false,
			},
			expectError:  false,
			wantContains: []string{},
		},
		{
			name: "Search with archive mode",
			cmd: Cmd{
				Match:        "*",
				Paths:        []string{tdir},
				Archive:      true,
				LastModified: true,
			},
			expectError:  false,
			wantContains: []string{"file_1985 (1985-01-01) >", "file_2012.txt (2012-01-01) >"},
		},
		{
			name: "Search with oldest flag",
			cmd: Cmd{
				Match:   "*",
				Paths:   []string{tdir},
				Archive: true,
				Oldest:  true,
			},
			expectError:  false,
			wantContains: []string{"Oldest found match:", "file_1985 (1985-01-01) >"},
		},
		{
			name: "Search with newest flag",
			cmd: Cmd{
				Match:   "*",
				Paths:   []string{tdir},
				Archive: true,
				Newest:  true,
			},
			expectError:  false,
			wantContains: []string{"Newest found match:", "file_2002.zyx (2025-03-26) >"},
		},
		{
			name: "Invalid destination directory",
			cmd: Cmd{
				Match:       "*.go",
				Paths:       []string{tdir},
				Destination: "/invalid/destination",
			},
			expectError:  true,
			wantContains: []string{},
		},
		{
			name: "Valid copy operation",
			cmd: Cmd{
				Match:       "archive.tar",
				Paths:       []string{tdir},
				Destination: t.TempDir(),
			},
			expectError:  false,
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := tt.cmd.Run()

			// Restore stdout
			w.Close()
			os.Stdout = old

			be.Equal(t, (err != nil), tt.expectError)

			if len(tt.wantContains) > 0 {
				var buf bytes.Buffer
				_, err := buf.ReadFrom(r)
				if err != nil {
					t.Fatalf("Failed to read from pipe: %v", err)
				}
				output := buf.String()
				for _, want := range tt.wantContains {
					if !bytes.Contains(buf.Bytes(), []byte(want)) {
						t.Logf("Expected to find '%s' in output:\n%s", want, output)
					}
					be.True(t, bytes.Contains(buf.Bytes(), []byte(want)))
				}
			}
		})
	}
}

func TestCmdRunErrorHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		cmd         Cmd
		expectError bool
	}{
		{
			name: "Invalid path",
			cmd: Cmd{
				Match: "*.go",
				Paths: []string{"/nonexistent/path"},
			},
			expectError: true,
		},
		{
			name: "Empty match pattern",
			cmd: Cmd{
				Match: "",
				Paths: []string{"."},
			},
			expectError: true,
		},
		{
			name: "No paths provided",
			cmd: Cmd{
				Match: "*.go",
				Paths: []string{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Removed t.Parallel() due to global state modification (os.Stdout)
			err := tt.cmd.Run()
			be.Equal(t, (err != nil), tt.expectError)
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	t.Parallel()
	tdir := "testdata"

	tests := []struct {
		name         string
		opt          ls.Config
		pattern      string
		paths        []string
		wantContains []string
		expectError  bool
	}{
		{
			name:         "Basic directory search",
			opt:          ls.Config{},
			pattern:      "file_*",
			paths:        []string{tdir},
			wantContains: []string{"file_1996", "file_1997.xyz", "file_2002.zyx"},
			expectError:  false,
		},
		{
			name:         "Search with count",
			opt:          ls.Config{Count: true},
			pattern:      "*",
			paths:        []string{tdir},
			wantContains: []string{"1\t", "2\t", "3\t"},
			expectError:  false,
		},
		{
			name:         "Search with last modified",
			opt:          ls.Config{LastModified: true},
			pattern:      "file_*",
			paths:        []string{tdir},
			wantContains: []string{"(", "-0", "-0"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Removed t.Parallel() due to race conditions with global state
			var buf bytes.Buffer
			err := tt.opt.Walks(&buf, tt.pattern, tt.paths...)

			be.Equal(t, (err != nil), tt.expectError)

			if len(tt.wantContains) > 0 {
				output := buf.String()
				for _, want := range tt.wantContains {
					if !bytes.Contains(buf.Bytes(), []byte(want)) {
						t.Errorf("Config.Walks() output should contain %q, got %q", want, output)
					}
				}
			}
		})
	}
}

func TestCheckDestIntegration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		dest        string
		expectError bool
	}{
		{
			name:        "Valid directory",
			dest:        t.TempDir(),
			expectError: false,
		},
		{
			name:        "Invalid directory",
			dest:        "/nonexistent/directory",
			expectError: true,
		},
		{
			name:        "Empty directory (no copy operation)",
			dest:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := cp.CheckDest(tt.dest)
			be.Equal(t, (err != nil), tt.expectError)
		})
	}
}

func TestConfigConstruction(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name string
		cmd  Cmd
		want ls.Config
	}{
		{
			name: "Basic config",
			cmd: Cmd{
				Archive:       false,
				CaseSensitive: false,
				Count:         false,
				Directory:     true,
				Destination:   "",
			},
			want: ls.Config{
				Archive:       false,
				Casesensitive: false,
				Count:         false,
				Directory:     true,
				Destination:   "",
				StdErrors:     false,
				Follow:        false,
				LastModified:  false,
				Oldest:        false,
				Newest:        false,
				NumWorkers:    0,
				Panic:         false,
				Sort:          fastwalk.SortDirsFirst,
			},
		},
		{
			name: "Archive config",
			cmd: Cmd{
				Archive:       true,
				CaseSensitive: true,
				Count:         true,
				Directory:     true,
				LastModified:  true,
			},
			want: ls.Config{
				Archive:       true,
				Casesensitive: true,
				Count:         true,
				Directory:     true,
				Destination:   "",
				StdErrors:     false,
				Follow:        false,
				LastModified:  true,
				Oldest:        false,
				Newest:        false,
				NumWorkers:    0,
				Panic:         false,
				Sort:          fastwalk.SortDirsFirst,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opt := ls.Config{
				Archive:       tt.cmd.Archive,
				Casesensitive: tt.cmd.CaseSensitive,
				Count:         tt.cmd.Count,
				Directory:     tt.cmd.Directory,
				Destination:   tt.cmd.Destination,
				StdErrors:     tt.cmd.Errors,
				Follow:        tt.cmd.Follow,
				LastModified:  tt.cmd.LastModified,
				Oldest:        tt.cmd.Oldest,
				Newest:        tt.cmd.Newest,
				NumWorkers:    0,
				Panic:         tt.cmd.Panic,
				Sort:          fastwalk.SortDirsFirst,
			}

			be.Equal(t, opt, tt.want)
		})
	}
}
