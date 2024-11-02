package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/bengarrett/namezed/ls"
	"github.com/charlievieth/fastwalk"
)

// ls, cp, lsz, cpz

type LsCmd struct {
	CaseSensitive bool     `help:"Case sensitive match." short:"c"`
	Count         bool     `help:"Count the number of matches." short:"n"`
	Directories   bool     `help:"Include directory matches." short:"d" default:"true"`
	Errors        bool     `help:"Errors mode displays any file and directory read or access errors." short:"e"`
	Follow        bool     `help:"Follow symbolic links." short:"f"`
	Panic         bool     `help:"Exits on any errors including file and directory read or access errors." short:"p"`
	Workers       int      `help:"Number of workers to use or leave it to the app." short:"w" default:"0"`
	Match         string   `arg:"" optional:"" help:"Filename, extension or pattern to match."`
	Paths         []string `arg:"" optional:"" help:"Paths to lookup." type:"path"`
}

func (cmd *LsCmd) Run() error {
	opt := ls.Config{
		Casesensitive: cmd.CaseSensitive,
		Count:         cmd.Count,
		Directories:   cmd.Directories,
		StdErrors:     cmd.Errors,
		Follow:        cmd.Follow,
		NumWorkers:    0,
		Panic:         cmd.Panic,
		Sort:          fastwalk.SortDirsFirst,
	}
	return opt.Walks(os.Stdout, cmd.Match, cmd.Paths...)
}

var cli struct {
	Ls LsCmd `cmd:"" help:"List the matching files."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("namzd"),
		kong.Description("Quickly find files by name or extension."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
