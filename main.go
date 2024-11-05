package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/bengarrett/namezed/cp"
	"github.com/bengarrett/namezed/ls"
	"github.com/charlievieth/fastwalk"
)

// LsCmd is the command line options for the ls command.
type LsCmd struct {
	Archive       bool     `help:"Archive mode will also search supported archives for matched files." short:"a"`
	CaseSensitive bool     `help:"Case sensitive match." short:"c"`
	Count         bool     `help:"Count the number of matches." short:"n"`
	Directory     bool     `help:"Include directory matches." short:"d" default:"true"`
	Errors        bool     `help:"Errors mode displays any file and directory read or access errors." short:"e"`
	Follow        bool     `help:"Follow symbolic links." short:"f"`
	Panic         bool     `help:"Exits on any errors including file and directory read or access errors." short:"p"`
	Destination   string   `help:"Destination directory path to copy matches." type:"path" short:"x"`
	Worker        int      `help:"Number of workers to use or leave it to the app." short:"w" default:"0"`
	Match         string   `arg:"" required:"" help:"Filename, extension or pattern to match."`
	Paths         []string `arg:"" required:"" help:"Paths to lookup." type:"path"`
}

// Run the ls command.
func (cmd *LsCmd) Run() error {
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
