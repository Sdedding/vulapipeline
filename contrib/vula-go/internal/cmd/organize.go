package cmd

import (
	"context"
	"os"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/filerepo"
	"codeberg.org/vula/vula/contrib/vula-go/internal/organize"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vuladbus"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vulanetlink"
)

type OrganizeCmd struct {
	Table          int    `short:"t" help:"Which routing table to use" default:"666"`
	Interface      string `short:"I" help:"WireGuard interface name" default:"vula"`
	StateFile      string `short:"c" help:"YAML state file" default:"/var/lib/vula-organize/vula-organize.yaml"`
	KeysFile       string `short:"k" help:"YAML configuration file for cryptographic keys" default:"/var/lib/vula-organize/keys.yaml"`
	Port           uint16 `short:"p" help:"path to base directory for organize state" default:"5354"`
	Fwmark         int    `short:"m" help:"path to base directory for organize state" default:"555"`
	IpRulePriority int    `short:"r" help:"path to base directory for organize state" default:"666"`
	// TODO: subcommands
}

func (command *OrganizeCmd) Run(g *Globals) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if g.Verbose {
		core.SetDebugLogOutput(os.Stderr)
	}

	err := runOrganizeDbus(ctx, command)
	return err
}

func runOrganizeDbus(ctx context.Context, c *OrganizeCmd) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	keyFileRepository := &filerepo.KeyFile{FileName: c.KeysFile}
	hostsFileRepository := &filerepo.HostsFile{FileName: core.HostsFileName}
	stateFileRepository := &filerepo.OrganizeStateFile{FileName: c.StateFile}

	discoverClient, err := vuladbus.NewDiscoverClient()
	if err != nil {
		return err
	}

	publishClient, err := vuladbus.NewPublishClient()
	if err != nil {
		return err
	}

	config := &organize.Config{
		StateRepository:     stateFileRepository,
		HostsFileRepository: hostsFileRepository,
		KeyRepository:       keyFileRepository,
		Discover:            discoverClient,
		Publish:             publishClient,
		NetworkSystem:       &vulanetlink.NetworkSystem{},
		WireguardSystem:     &vulanetlink.WireguardSystem{},

		InterfaceName:  c.Interface,
		Port:           uint16(c.Port),
		FirewallMark:   c.Fwmark,
		RoutingTable:   c.Table,
		IPRulePriority: c.IpRulePriority,
	}

	o, err := organize.New(config)
	if err != nil {
		return err
	}
	defer util.CloseLog(o)

	var (
		dbusServerErr = make(chan error)
	)

	go func() {
		dbusServerErr <- vuladbus.RunOrganizeServer(ctx.Done(), o)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-dbusServerErr:
		return err
	}
}
