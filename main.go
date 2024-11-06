package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bengarrett/namzd/cp"
	"github.com/bengarrett/namzd/ls"
	"github.com/charlievieth/fastwalk"
)

type Globals struct {
	Version VersionFlag `name:"version" help:"Show the version information and exit." short:"V"`
}

// VersionFlag is a custom flag type for the display of the version information.
type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

// Cmd is the command line options for the ls command.
type Cmd struct {
	Archive       bool     `group:"zip" xor:"x0" help:"Archive mode will also search within supported archives for matched filenames." short:"a"`
	Destination   string   `group:"copy" xor:"x0,x1" help:"Destination directory path to copy matches." type:"existingdir" short:"x"`
	CaseSensitive bool     `help:"Case sensitive match." short:"c"`
	Count         bool     `help:"Count the number of matches." short:"n"`
	Directory     bool     `xor:"x1" help:"Include directory matches." short:"d" default:"true"`
	Errors        bool     `group:"errs" help:"Errors mode displays any file and directory read or access errors." short:"e"`
	Follow        bool     `help:"Follow symbolic links." short:"f"`
	Panic         bool     `group:"errs" help:"Exits on any errors including file and directory read or access errors." short:"p"`
	Worker        int      `help:"Number of workers to use or leave it to the app." short:"w" default:"0" hidden:""`
	Match         string   `arg:"" required:"" help:"Filename, extension or pattern to match."`
	Paths         []string `arg:"" required:"" help:"Paths to lookup." type:"existingdir"`
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
		NumWorkers:    0,
		Panic:         cmd.Panic,
		Sort:          fastwalk.SortDirsFirst,
	}
	if err := cp.CheckDest(opt.Destination); err != nil {
		return err
	}

	return opt.Walks(os.Stdout, cmd.Match, cmd.Paths...)
}

func help() string {
	s := `

A <match> query is a filename, extension or pattern to match.
These are case-insensitive by default and should be quoted:

	'readme' matches README, Readme, readme, etc.
	'file.txt' matches file.txt, File.txt, file.TXT, etc.
	'*.txt' matches readme.txt, File.txt, DOC.TXT, etc.
	'*.tar*' matches files.tar.gz, FILE.tarball, files.tar, files.tar.xz, etc.
	'*.tar.??' matches files.tar.gz, files.tar.xz, etc.`
	return s
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
		Description: "Copy all matched files to a target directory. This option cannot be used with the archive options or the directory flag.",
	}
	errgrp := kong.Group{
		Key:   "errs",
		Title: "Errors:",
	}
	zipgrp := kong.Group{
		Key:         "zip",
		Title:       "Archives:",
		Description: "Also search within archives for matching files. This will not open or decompress archives to read archives within archives.",
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
