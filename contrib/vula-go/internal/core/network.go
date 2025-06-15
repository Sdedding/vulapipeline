package core

import "net/netip"

type NetworkSystem interface {
	// StartMonitor starts monitoring network interfaces
	StartMonitor(done <-chan struct{}, updates chan<- string) error

	// GetSystemState reads the current system state
	GetSystemState(enableIPv4, enableIPv6 bool, ifacePrefixAllowed []string, subnetsForbidden []netip.Prefix, primaryIP netip.Addr) (NetworkSystemState, error)

	// SyncRules creates ip rules if they dont exist
	SyncRules(routingTable, firewallMark, priority int, dryRun bool) ([]string, error)

	// SyncRoutes creates routes if they dont exist
	SyncRoutes(dests []netip.Prefix, routingTable int, interfaceName string, currentSubnets map[netip.Prefix][]netip.Addr, dryRun bool) ([]string, error)

	// RemoveRoutes removes the specified routes
	RemoveRoutes(dests []netip.Prefix, routingTable int, interfaceName string, dryRun bool) ([]string, error)

	// RemoveUnknownRoutes removes the routes if they are not in expected routes
	RemoveUnknownRoutes(expectedRoutes map[netip.Prefix]struct{}, routingTable int, hasEnabledGateway, dryRun bool) ([]string, error)
}

type NetworkSystemState struct {
	CurrentSubnets    map[netip.Prefix][]netip.Addr
	CurrentInterfaces map[string][]netip.Addr
	Gateways          []netip.Addr
	HasV6             bool
}
