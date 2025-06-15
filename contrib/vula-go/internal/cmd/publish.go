package cmd

import (
	"context"
	"os"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vuladbus"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vulazeroconf"
)

type PublishCmd struct {
}

func (command *PublishCmd) Run(g *Globals) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if g.Verbose {
		core.SetDebugLogOutput(os.Stderr)
	}

	err := runPublishDbus(ctx)
	return err
}

func runPublishDbus(ctx context.Context) error {
	organizeClient, err := vuladbus.NewOrganizeClient()
	if err != nil {
		return err
	}
	defer util.CloseLog(organizeClient)

	publish := vulazeroconf.NewPublish()
	defer util.CloseLog(publish)

	return vuladbus.RunPublishServer(ctx.Done(), publish)
}
