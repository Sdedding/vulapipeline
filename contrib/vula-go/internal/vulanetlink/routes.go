package vulanetlink

import (
	"fmt"
	"net"
	"net/netip"
	"slices"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/vishvananda/netlink"
)

// removeRoutes idempotently remove route(s)
func removeRoutes(dests map[netip.Prefix]struct{}, routingTable int, interfaceName string, dryRun bool) (log []string, err error) {
	routeEntries, err := getRouteEntries(dests, &routingTable, interfaceName)
	if err != nil {
		return
	}

	for i := range routeEntries {
		route := &routeEntries[i]
		if !dryRun {
			err = netlink.RouteDel(route)
			if err != nil {
				return
			}
		}
		log = append(log, fmt.Sprintf("ip route del %s dev %s table %d", route.Dst, interfaceName, routingTable))
	}
	return
}

// getRouteEntries queries for routes
func getRouteEntries(dests map[netip.Prefix]struct{}, routingTable *int, device string) ([]netlink.Route, error) {
	currentRoutes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(device)
	if err != nil {
		return nil, err
	}

	filteredRoutes := []netlink.Route{}
	for _, route := range currentRoutes {
		prefix, err := ipNetToPrefix(route.Dst)
		if err != nil {
			return filteredRoutes, err
		}
		prefix = prefix.Masked()
		_, inDests := dests[prefix]
		keep := (dests == nil || inDests) &&
			(routingTable == nil || route.Table == *routingTable) &&
			route.LinkIndex == link.Attrs().Index

		if keep {
			filteredRoutes = append(filteredRoutes, route)
		}
	}

	return filteredRoutes, nil
}

func removeUnknownRoutes(expectedRoutes map[netip.Prefix]struct{}, routingTable int, hasEnabledGateway, dryRun bool) (log []string, err error) {
	currentRoutes, err := netlink.RouteList(nil, 0)
	if err != nil {
		return
	}

	// remove peer routes
	for i := range currentRoutes {
		route := &currentRoutes[i]

		if route.Table != routingTable {
			continue
		}
		var dst netip.Prefix
		dst, err = ipNetToPrefix(route.Dst)
		if err != nil {
			return
		}
		dst = dst.Masked()

		if _, ok := expectedRoutes[dst]; !ok {
			if !dryRun {
				core.LogInfof("Removing unexpected route: (%s)", dst)
				err = netlink.RouteDel(route)
				if err != nil {
					err = fmt.Errorf("%w on ip route del %s table %d scope %v", err, dst, routingTable, route.Scope)
					return
				}
			}

			log = append(log, fmt.Sprintf("ip route del %s table %d scope %v", dst, routingTable, route.Scope))
		}
	}

	if !hasEnabledGateway {
		for i := range currentRoutes {
			route := &currentRoutes[i]
			var dst netip.Prefix
			dst, err = ipNetToPrefix(route.Dst)
			if err != nil {
				return
			}
			dst = dst.Masked()

			isDefaultRoute := route.Table == core.LinuxMainRoutingTable && slices.Contains(core.GatewayRoutes, dst)
			if !isDefaultRoute {
				continue
			}

			if !dryRun {
				core.LogInfof("Removing unexpected route: (%s)", dst)
				err = netlink.RouteDel(route)
				if err != nil {
					err = fmt.Errorf("error deleting gateway route: %w", err)
					return
				}
			}

			log = append(log, fmt.Sprintf("ip route del %s table %d scope %v", dst, core.LinuxMainRoutingTable, route.Scope))
		}
	}
	return
}

// syncRoutes takes a list of cidr notation dests and a routing table, and ensures
// those routes are configured there. Returns a string.
func syncRoutes(dests []netip.Prefix, routingTable int, interfaceName string, currentSubnets map[netip.Prefix][]netip.Addr, dryRun bool) (log []string, err error) {
	core.LogDebugf("looking for routes for: %s", dests)

	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return
	}

	allRoutes, err := netlink.RouteList(nil, 0)
	if err != nil {
		return log, err
	}

	for _, dest := range dests {
		routes := []netlink.Route{}
		for _, route := range allRoutes {
			dst, err := ipNetToPrefix(route.Dst)
			if err != nil {
				return nil, err
			}
			dst = dst.Masked()
			if dst == dest && route.Table == routingTable {
				routes = append(routes, route)
			}
		}

		if len(routes) == 0 {
			var src *netip.Addr
			for net := range currentSubnets {
				/*
				 * note: current_subnets is consulted to find a source
				 * address but NOT consulted regarding the destination.
				 * (for pinned peers, we want to add IPs from
				 * non-current subnets here; they only need to be in a
				 * current subnet the first time they're seen)
				 */

				if net.Contains(dest.Addr()) {
					/*
					 * select the first local IP we have in the first
					 * subnet. (it would be more correct to use the
					 * longest-prefix-matching subnet... but not
					 * bothering with that for now.)
					 */

					src = &currentSubnets[net][0]
					break

					/*
					 * further note: this current_subnets logic
					 * doesn't belong here at all; will refactor
					 * this into system state soon.
					 */
				}
			}

			srcString := ""
			if src != nil {
				srcString = fmt.Sprintf("src %s", src.String())
			}
			log = append(log, fmt.Sprintf("ip route add %s dev %s proto static scope link %s table %d", dest, interfaceName, srcString, routingTable))

			if !dryRun {
				var netSrc net.IP
				if src != nil {
					netSrc = net.IP(src.AsSlice())
				}

				core.LogInfof("[#] %s", log[len(log)-1])
				err = netlink.RouteAdd(&netlink.Route{
					Dst: &net.IPNet{
						IP:   net.IP(dest.Addr().AsSlice()),
						Mask: net.CIDRMask(dest.Addr().BitLen(), dest.Addr().BitLen()),
					},
					LinkIndex: link.Attrs().Index,
					Table:     routingTable,
					Scope:     netlink.SCOPE_LINK,
					Src:       netSrc,
				})
				if err != nil {
					return log, err
				}
			}
		} else {
			for _, route := range routes {
				core.LogDebugf("found existing route for %s: %s", dest, route)
			}
		}
	}
	return
}
