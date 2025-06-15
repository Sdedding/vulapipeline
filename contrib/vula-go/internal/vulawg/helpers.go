package vulawg

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/netip"
)

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

type errInvalidNetlinkPrefix struct {
	net *net.IPNet
}

func (e *errInvalidNetlinkPrefix) Error() string {
	return fmt.Sprintf("invalid ip net from wgctrl: %s", e.net)
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
