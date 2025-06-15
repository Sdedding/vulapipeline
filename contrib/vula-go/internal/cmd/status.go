package cmd

import (
	"fmt"
	"os"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/organize"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vuladbus"
)

type StatusCmd struct {
	OnlySystemd bool `short:"s" help:"Only print systemd service status"`
}

func (c *StatusCmd) Help() string {
	return `Print status of systemd services and system configuration`
}

func (command *StatusCmd) Run(g *Globals) error {
	if g.Verbose {
		core.SetDebugLogOutput(os.Stderr)
	}
	systemd, err := vuladbus.NewSystemdClient()
	if err != nil {
		return err
	}

	meta, err := vuladbus.NewMetaClient()
	if err != nil {
		return err
	}

	status, err := organize.Status(systemd, meta)
	if err != nil {
		return err
	}
	fmt.Println(string(status.ToConsole()))
	return nil
}
