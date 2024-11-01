package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
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
		conf := fastwalk.Config{
			Follow: true,
		}

		walkFn := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				return nil // returning the error stops iteration
			}
			if cli.Ls.Name != "" {
				pattern := strings.ToLower(cli.Ls.Name)
				name := strings.ToLower(d.Name()) // basename
				if ok, err := filepath.Match(pattern, name); !ok {
					// invalid pattern (err != nil) or name does not match
					return err
				}
			}
			_, err = fmt.Println(path)
			return err
		}
		root := cli.Ls.Paths[0]
		if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
			os.Exit(1)
		}

	default:
		fmt.Println("unknown command:", ctx.Command())
	}
}
