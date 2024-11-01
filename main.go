package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/bengarrett/namezed/ls"
	"github.com/charlievieth/fastwalk"
)

var cli struct {
	Debug bool `help:"Debug mode."`

	Rm struct {
		User      string `help:"Run as user." short:"u" default:"default"`
		Force     bool   `help:"Force removal." short:"f"`
		Recursive bool   `help:"Recursively remove files." short:"r"`

		Paths []string `arg:"" help:"Paths to remove." type:"path" name:"path"`
	} `cmd:"" help:"Remove files."`

	Ls struct {
		Name  string   `arg:"" optional:"" help:"Filename to match."`
		Paths []string `arg:"" optional:"" help:"Paths to list." type:"path"`
	} `cmd:"" help:"List paths."`
}

// ls, cp, lsz, cpz

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("shell"),
		kong.Description("A shell-like example app."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	switch ctx.Command() {
	case "rm <path>":
		fmt.Println(cli.Rm.Paths, cli.Rm.Force, cli.Rm.Recursive)

	case "ls <name> <paths>":
		fmt.Println(">", cli.Ls.Name, cli.Ls.Paths)

		// Follow links if the "-L" flag is provided
		opt := ls.Config{
			Casesensitive: false,
			Count:         true,
			Directories:   false,
			Follow:        true,
			NumWorkers:    0,
			Sort:          fastwalk.SortDirsFirst,
		}
		opt.Walks(cli.Ls.Name, cli.Ls.Paths...)

	default:
		fmt.Println("unknown command:", ctx.Command())
	}
}
