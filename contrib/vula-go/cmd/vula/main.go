package main

import (
	"codeberg.org/vula/vula/contrib/vula-go/internal/cmd"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"fmt"
	"github.com/alecthomas/kong"
	"os"
)

var version string

type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	cmd.Globals
	Version VersionFlag `help:"Show the version and exit."`

	Discover cmd.DiscoverCmd `cmd:"" help:"Layer 3 mDNS discovery daemon"`
	Organize cmd.OrganizeCmd `cmd:"" help:"Maintain routes and wg peer configurations"`
	Peer     cmd.PeerCmd     `cmd:"" help:"View and modify peer information"`
	Publish  cmd.PublishCmd  `cmd:"" help:"Layer 3 mDNS publish daemon"`
	Status   cmd.StatusCmd   `cmd:"" help:"Print status"`
}

func run() {
	cli := CLI{
		Globals: cmd.Globals{},
		Version: VersionFlag(version),
	}
	if len(os.Args) < 2 {
		// The status command is the default command
		os.Args = append(os.Args, "status")

		// If just `vula` is invoked, it runs the status subcommand, so essentially `vula status`.
		// However, this command has the optional `-s` flag. Note however that running `vula -s` will not do the
		// equivalent of running `vula status -s`.
	}
	ctx := kong.Parse(&cli,
		kong.Name("vula"),
		kong.Description("vula tools\n\nWith no arguments, runs \"status\""),
		kong.UsageOnError(),
		kong.Help(util.VulaHelpPrinter),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			NoExpandSubcommands: true,
		}),
	)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}

func main() {
	run()
}
