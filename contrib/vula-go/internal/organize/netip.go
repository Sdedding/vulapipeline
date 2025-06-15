package organize

import (
	"net/netip"
	"slices"
)

// sortLLFirst sorts a list of IPs to put the link-local ones (if any) first, and to
// secondarily to place v6 addresses ahead of v4.
func sortLLFirst(addrs []netip.Addr) {
	slices.SortStableFunc(addrs, func(a, b netip.Addr) int {
		aIsLinkLocal := a.IsLinkLocalUnicast()
		bIsLinkLocal := b.IsLinkLocalUnicast()

		if aIsLinkLocal && !bIsLinkLocal {
			return -1
		}
		if !aIsLinkLocal && bIsLinkLocal {
			return 1
		}
		aIs6 := a.Is6()
		bIs6 := b.Is6()

		if aIs6 && !bIs6 {
			return -1
		}
		if !aIs6 && bIs6 {
			return 1
		}
		return 0
	})
}

func addrsInSubnets(addrs []netip.Addr, subnets map[netip.Prefix][]netip.Addr) []netip.Addr {
	s := []netip.Addr{}
	for _, addr := range addrs {
		for subnet := range subnets {
			if subnet.Contains(addr) {
				s = append(s, addr)
				break
			}
		}
	}
	return s
}
