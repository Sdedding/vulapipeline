package vulanetlink

import (
	"fmt"
	"net"
	"net/netip"
)

type errInvalidNetlinkPrefix struct {
	net *net.IPNet
}

func (e *errInvalidNetlinkPrefix) Error() string {
	return fmt.Sprintf("invalid ip net from netlink: %s", e.net)
}

func ipNetToPrefix(ipNet *net.IPNet) (prefix netip.Prefix, err error) {
	if ipNet == nil {
		err = &errInvalidNetlinkPrefix{ipNet}
	}
	addr, ok := netip.AddrFromSlice(ipNet.IP)
	if !ok {
		err = &errInvalidNetlinkPrefix{ipNet}
		return
	}
	ones, bits := ipNet.Mask.Size()
	if bits == 0 {
		err = &errInvalidNetlinkPrefix{ipNet}
		return
	}
	prefix = netip.PrefixFrom(addr, ones)
	return
}
