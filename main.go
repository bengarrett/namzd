// Package main implements the namzd CLI tool for quickly finding files by name or extension across directories and archives.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bengarrett/namzd/cp"
	"github.com/bengarrett/namzd/ls"
	"github.com/charlievieth/fastwalk"
)

var (
	ErrMatchRequired = errors.New("match pattern is required")
	ErrPathsRequired = errors.New("at least one path is required")
)

type Globals struct {
	Version VersionFlag `help:"Show the version information and exit." name:"version" short:"V"`
}

// VersionFlag is a custom flag type for the display of the version information.
type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error { //nolint:unparam
	fmt.Fprintln(os.Stdout, vars["version"])
	app.Exit(0)
	return nil
}

// Cmd is the command line options for the ls command.
type Cmd struct {
	Archive       bool     `group:"zip"                                                   help:"Archive mode will also search within supported archives (ZIP, TAR, TAR.GZ, TAR.XZ)."  short:"a"   xor:"x0"`
	Destination   string   `group:"copy"                                                  help:"Destination directory path to copy matched files (cannot be used with archive mode)." short:"x"   type:"existingdir" xor:"x0"`
	CaseSensitive bool     `help:"Case sensitive match."                                  short:"c"`
	Count         bool     `help:"Count the number of matches."                           short:"n"`
	LastModified  bool     `help:"Show the last modified time of the match (yyyy-mm-dd)." short:"m"`
	Oldest        bool     `help:"Show the oldest file match."                            short:"o"`
	Newest        bool     `help:"Show the newest file match."                            short:"N"`
	Directory     bool     `default:"true"                                                help:"Include directory matches."                                                           short:"d"   xor:"x1"`
	Errors        bool     `group:"errs"                                                  help:"Errors mode displays any file and directory read or access errors."                   short:"e"`
	Follow        bool     `help:"Follow symbolic links."                                 short:"f"`
	NoColor       bool     `help:"No color output is faster."                             short:"C"`
	Panic         bool     `group:"errs"                                                  help:"Exits on any errors including file and directory read or access errors."              short:"p"`
	Worker        int      `default:"0"                                                   help:"Number of workers to use or leave it to the app."                                     hidden:""   short:"w"`
	Match         string   `arg:""                                                        help:"Filename, extension or pattern to match."                                             required:""`
	Paths         []string `arg:""                                                        help:"Paths to lookup."                                                                     required:"" type:"existingdir"`
}

// Run the ls command.
func (cmd *Cmd) Run() error {
	opt := ls.Config{
		Archive:       cmd.Archive,
		Casesensitive: cmd.CaseSensitive,
		Count:         cmd.Count,
		Directory:     cmd.Directory,
		Destination:   cmd.Destination,
		StdErrors:     cmd.Errors,
		Follow:        cmd.Follow,
		LastModified:  cmd.LastModified,
		Oldest:        cmd.Oldest,
		Newest:        cmd.Newest,
		NoColor:       cmd.NoColor,
		NumWorkers:    0,
		Panic:         cmd.Panic,
		Sort:          fastwalk.SortDirsFirst,
	}

	// Validate required arguments
	if cmd.Match == "" {
		return fmt.Errorf("run cmd: %w", ErrMatchRequired)
	}
	if len(cmd.Paths) == 0 {
		return fmt.Errorf("run cmd: %w", ErrPathsRequired)
	}

	if err := cp.CheckDest(opt.Destination); err != nil {
		return fmt.Errorf("run ls: %w", err)
	}
	if err := opt.Walks(os.Stdout, cmd.Match, cmd.Paths...); err != nil {
		return fmt.Errorf("run ls: %w", err)
	}
	return nil
}

func help() string {
	help := `

A <match> query is a filename, extension or pattern.
These are case-insensitive by default and should be quoted:

	'readme'   matches README, Readme, readme, etc.
	'file.txt' matches file.txt, File.txt, file.TXT, etc.
	'*.txt'    matches readme.txt, File.txt, DOC.TXT, etc.
	'*.tar*'   matches files.tar.gz, FILE.tarball, files.tar, files.tar.xz, etc.
	'*.tar.??' matches files.tar.gz, files.tar.xz, etc.

	Always quote wildcard patterns:

	namzd '*.txt' /path  # correct, namzd processes the wildcard
	namzd  *.txt  /path  # invalid, as the shell expands *.txt first

Examples:

	# Find all README files in current directory
	namzd 'readme' .

	# Find all Go files in a code directory
	namzd '*.go' /path/to/code

	# Case-sensitive search for config files
	namzd 'config' /etc --case-sensitive

	# Count text files in documents
	namzd '*.txt' /documents --count

	# Find oldest backup file
	namzd 'backup' /archives --oldest

Flag Incompatibility:

	--archive and --destination cannot be used together`
	return help
}

type CLI struct {
	Globals
	Cmd `cmd:""`
}

var (
	version = "development"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli := CLI{
		Globals: Globals{
			Version: VersionFlag("development"),
		},
	}

	cpgrp := kong.Group{
		Key:         "copy",
		Title:       "Copier:",
		Description: "Copy all matched files to a target directory. This option cannot be used with the archive options.",
	}
	errgrp := kong.Group{
		Key:   "errs",
		Title: "Errors:",
	}
	zipgrp := kong.Group{
		Key:         "zip",
		Title:       "Archives:",
		Description: "Also search within archives for matching files. This will not recursively search archives contained within archives.",
	}

	ctx := kong.Parse(&cli,
		kong.Name("namzd"),
		kong.Description("Quickly find files by name or extension."+help()),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: false,
			Compact: false,
		}),
		kong.ExplicitGroups([]kong.Group{cpgrp, errgrp, zipgrp}),
		kong.Vars{
			"version": fmt.Sprintf("namzd - Quickly find files.\n"+
				"%s (commit %s), built at %s.\n%s",
				version, commit, date, "https://github.com/bengarrett/namzd"),
		})
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
