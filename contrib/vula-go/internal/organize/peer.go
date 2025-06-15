package organize

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"slices"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

// peerID returns the peer ID (aka the verify key, as a base64 string)
func peerID(p *core.Peer) string {
	return base64.StdEncoding.EncodeToString(p.Descriptor.VulaPK)
}

// peerName returns the best name, for display purposes
func peerName(p *core.Peer) string {
	if p.Petname != "" {
		return p.Petname
	}

	if p.Nicknames[p.Descriptor.Hostname] {
		return p.Descriptor.Hostname
	}

	enabledNames := peerEnabledNames(p)
	if len(enabledNames) > 0 {
		return enabledNames[0]
	}

	return "<unnamed>"
}

func peerEnabledNames(p *core.Peer) []string {
	m := map[string]struct{}{}
	if p.Petname != "" {
		m[p.Petname] = struct{}{}
	}

	for name, enabled := range p.Nicknames {
		if enabled {
			m[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	slices.Sort(names)
	return names
}

// peerNameAndID returns the name and id of the peer
func peerNameAndID(p *core.Peer) string {
	return fmt.Sprintf("%s (%s)", peerName(p), peerID(p))
}

func peerEndpointAddr(p *core.Peer) netip.Addr {
	if len(p.Addrs) == 0 {
		return netip.AddrFrom4([4]byte{})
	}

	addrs := []netip.Addr{}
	for _, addr := range peerEnabledIPs(p, true) {
		if !core.VulaSubnet.Contains(addr) {
			addrs = append(addrs, addr)
		}
	}
	sortLLFirst(addrs)
	return addrs[0]
}

func peerEnabledIPs(p *core.Peer, enabled bool) []netip.Addr {
	var ips []netip.Addr
	if enabled {
		ips = append(ips, p.Descriptor.PrimaryIP)
	}
	for ip, on := range p.Addrs {
		if on == enabled {
			ips = append(ips, ip)
		}
	}
	return ips
}

func peerRoutableIPs(p *core.Peer) []netip.Addr {
	enabledIPs := peerEnabledIPs(p, true)
	addrs := make([]netip.Addr, 0, len(enabledIPs))
	for _, ip := range enabledIPs {
		if !(ip.Is6() && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast())) {
			addrs = append(addrs, ip)
		}
	}
	return addrs
}

func peerRoutes(p *core.Peer) []netip.Prefix {
	routableIPs := peerRoutableIPs(p)
	networks := make([]netip.Prefix, len(routableIPs))
	for i, ip := range routableIPs {
		networks[i], _ = ip.Prefix(ip.BitLen())
	}
	return networks
}

func peerWgAllowedIPs(p *core.Peer) []netip.Prefix {
	networks := peerRoutes(p)
	if p.UseAsGateway {
		networks = append(networks,
			netip.MustParsePrefix("0.0.0.0/0"),
			netip.MustParsePrefix("::/0"))
	}
	return networks
}

func peerWireGuardConfig(p *core.Peer, ctidhPsk []byte) *core.PeerConfig {
	return &core.PeerConfig{
		PublicKey:    p.Descriptor.WireGuardPK,
		EndpointAddr: peerEndpointAddr(p),
		EndpointPort: p.Descriptor.WireGuarPort,
		AllowedIPs:   peerWgAllowedIPs(p),
		PresharedKey: ctidhPsk,
	}
}

func peerEndpoint(p *core.Peer) string {
	addr := peerEndpointAddr(p)
	port := p.Descriptor.WireGuarPort
	if addr.Is6() {
		return fmt.Sprintf("[%s]:%d", addr, port)
	}
	return fmt.Sprintf("%s:%d", addr, port)
}

// peerOtherNames returns a sorted list of all names other than peerName
func peerOtherNames(p *core.Peer) []string {
	names := []string{}
	name := peerName(p)

	for n, enabled := range p.Nicknames {
		if n != name && enabled {
			names = append(names, n)
		}
	}

	slices.Sort(names)
	return names
}

func peerEnabledIPsString(p *core.Peer) []string {
	//TODO: possible IPv6 issue
	ips := peerEnabledIPs(p, true)
	s := make([]string, len(ips))
	for i := range ips {
		s[i] = ips[i].String()
	}
	return s
}
