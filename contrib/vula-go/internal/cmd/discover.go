package cmd

import (
	"context"
	"os"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vuladbus"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vulazeroconf"
)

type DiscoverCmd struct {
	// TODO: The option in the python version is "-d, --dbus / --no-dbus  use dbus for IPC". Is dbus used by default?
	Dbus bool `short:"d" negatable:"" help:"use dbus for IPC" default:"true"`
	// TODO: In the python version these help texts are wrapped in the second column.
	// We may want to do the same in VulaHelpPrinter.
	IpAddress string `short:"I" help:"bind this IP address instead of automatically choosing which IP to bind"`
	Interface string `help:"bind to the primary IP address for the given interface, automatically choosing which IP to announce"`
}

func (command *DiscoverCmd) Run(g *Globals) error {

	// TODO: handle the options

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if g.Verbose {
		core.SetDebugLogOutput(os.Stderr)
	}

	err := runDiscoverDbus(ctx)
	return err
}

func runDiscoverDbus(ctx context.Context) error {
	organizeClient, err := vuladbus.NewOrganizeClient()
	if err != nil {
		return err
	}
	defer util.CloseLog(organizeClient)

	discover, err := vulazeroconf.NewDiscover(organizeClient)
	if err != nil {
		return err
	}
	defer util.CloseLog(discover)

	err = vuladbus.RunDiscoverServer(ctx.Done(), discover)
	return err
}
