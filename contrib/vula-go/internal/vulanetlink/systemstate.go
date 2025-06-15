package vulanetlink

import (
	"net/netip"
	"slices"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/vishvananda/netlink"
)

func getNetlinkSystemState(enableIPv4, enableIPv6 bool, ifacePrefixAllowed []string, subnetsForbidden []netip.Prefix, primaryIP netip.Addr) (s core.NetworkSystemState, err error) {
	// wip

	addrs, err := getAllAddrs()
	if err != nil {
		return
	}

	gateways, err := getGateways()

	hasV6 := slices.ContainsFunc(addrs, func(addr interfaceAddr) bool {
		return addr.prefix.Addr().Is6()
	})

	currentSubnets := map[netip.Prefix][]netip.Addr{}
	currentInterfaces := map[string][]netip.Addr{}
	filteredAddrs := filterInterfaceAddrs(enableIPv4, enableIPv6, addrs, ifacePrefixAllowed, subnetsForbidden)

	for _, addr := range filteredAddrs {
		subnet := addr.prefix.Masked()

		addrsForSubnet := currentSubnets[subnet]
		addrsForSubnet = append(addrsForSubnet, addr.prefix.Addr())
		currentSubnets[subnet] = addrsForSubnet
		ifaceAddrs := currentInterfaces[addr.link]
		ifaceAddrs = append(ifaceAddrs, addr.prefix.Addr())
		currentInterfaces[addr.link] = ifaceAddrs
	}

	currentSubnets[core.VulaSubnet] = []netip.Addr{primaryIP}

	s.CurrentSubnets = currentSubnets
	s.CurrentInterfaces = currentInterfaces
	s.Gateways = gateways
	s.HasV6 = hasV6
	return
}

func getAllAddrs() (addrs []interfaceAddr, err error) {
	netlinkAddrs, err := netlink.AddrList(nil, 0)
	if err != nil {
		return
	}

	addrs = make([]interfaceAddr, len(netlinkAddrs))
	for i := range netlinkAddrs {
		var prefix netip.Prefix
		netlinkAddr := &netlinkAddrs[i]
		prefix, err = ipNetToPrefix(netlinkAddr.IPNet)
		if err != nil {
			return
		}

		var link netlink.Link
		link, err = netlink.LinkByIndex(netlinkAddr.LinkIndex)
		if err != nil {
			return
		}

		addrs[i] = interfaceAddr{link.Attrs().Name, prefix}
	}
	return
}

type interfaceAddr struct {
	link   string
	prefix netip.Prefix
}

func filterInterfaceAddrs(enableIPv4, enableIPv6 bool, addrs []interfaceAddr, ifacePrefixAllowed []string, subnetsForbidden []netip.Prefix) []interfaceAddr {
	filtered := []interfaceAddr{}
	for _, ifaceAddr := range addrs {
		if ifaceAddr.prefix.Addr().Is4() && !enableIPv4 {
			continue
		}
		if ifaceAddr.prefix.Addr().Is6() && !enableIPv6 {
			continue
		}
		hasAllowedPrefix := slices.ContainsFunc(ifacePrefixAllowed, func(prefix string) bool {
			return strings.HasPrefix(ifaceAddr.link, prefix)
		})
		if !hasAllowedPrefix {
			continue
		}
		isForbidden := slices.ContainsFunc(subnetsForbidden, func(forbiddenPrefix netip.Prefix) bool {
			return forbiddenPrefix.Contains(ifaceAddr.prefix.Addr())
		})
		if isForbidden {
			continue
		}

		filtered = append(filtered, ifaceAddr)
	}
	return filtered
}

func getGateways() ([]netip.Addr, error) {
	routes, err := netlink.RouteList(nil, 0)
	if err != nil {
		return nil, err
	}

	gateways := []netip.Addr{}
	for _, route := range routes {
		if len(route.Gw) > 0 {
			gw, ok := netip.AddrFromSlice(route.Gw)
			if !ok {
				core.LogDebug(gw)
				panic("vula: received invalid addr from netlink")
			}
			gateways = append(gateways, gw)
		}
	}
	return gateways, nil
}
