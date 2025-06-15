package vulanetlink

import (
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/vishvananda/netlink"
)

func startNetlinkMonitor(done <-chan struct{}, updates chan<- string) error {
	addrUpdates := make(chan netlink.AddrUpdate, 255)
	routeUpdates := make(chan netlink.RouteUpdate, 255)

	err := netlink.AddrSubscribe(addrUpdates, done)
	if err != nil {
		return err
	}

	err = netlink.RouteSubscribe(routeUpdates, done)
	if err != nil {
		return err
	}

	go func() {
		defer close(updates)

		for {
			select {
			case <-done:
				return
			case ev := <-addrUpdates:
				core.LogDebugf("acting on netlink event: %v", ev)
				updates <- "address"

			case ev := <-routeUpdates:
				core.LogDebugf("acting on netlink event: %s", ev)
				updates <- "route"
			}
		}
	}()

	return nil
}
