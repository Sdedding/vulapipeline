package vulanetlink

import (
	"net/netip"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/vulawg"
)

type NetworkSystem struct{}

var _ core.NetworkSystem = &NetworkSystem{}

func (*NetworkSystem) StartMonitor(done <-chan struct{}, updates chan<- string) error {
	return startNetlinkMonitor(done, updates)
}

func (*NetworkSystem) GetSystemState(enableIPv4, enableIPv6 bool, ifacePrefixAllowed []string, subnetsForbidden []netip.Prefix, primaryIP netip.Addr) (core.NetworkSystemState, error) {
	return getNetlinkSystemState(enableIPv4, enableIPv6, ifacePrefixAllowed, subnetsForbidden, primaryIP)
}

func (*NetworkSystem) SyncRules(routingTable, firewallMark, priority int, dryRun bool) ([]string, error) {
	return syncNetlinkIPRules(routingTable, firewallMark, priority, dryRun)
}

func (*NetworkSystem) SyncRoutes(dests []netip.Prefix, routeingTable int, interfaceName string, currentSubnets map[netip.Prefix][]netip.Addr, dryRun bool) ([]string, error) {
	return syncRoutes(dests, routeingTable, interfaceName, currentSubnets, dryRun)
}

func (*NetworkSystem) RemoveRoutes(dests []netip.Prefix, routingTable int, interfaceName string, dryRun bool) ([]string, error) {
	routes := make(map[netip.Prefix]struct{})
	for _, dest := range dests {
		routes[dest] = struct{}{}
	}

	return removeRoutes(routes, routingTable, interfaceName, dryRun)
}

func (*NetworkSystem) RemoveUnknownRoutes(expectedRoutes map[netip.Prefix]struct{}, routingTable int, hasEnabledGateway, dryRun bool) ([]string, error) {
	return removeUnknownRoutes(expectedRoutes, routingTable, hasEnabledGateway, dryRun)
}

type WireguardSystem struct{}

var _ core.WireguardSystem = &WireguardSystem{}

func (*WireguardSystem) GetInterface(name string) core.WireguardInterface {
	return vulawg.NewInterface(name)
}

func (*WireguardSystem) SyncInterface(name string, primaryIP netip.Addr, dryRun bool) ([]string, error) {
	return syncInterfaces(name, primaryIP, dryRun)
}
