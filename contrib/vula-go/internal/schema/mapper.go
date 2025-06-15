package schema

import (
	"fmt"
	"math"
	"net/netip"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func ImportDescriptor(s *Descriptor) (d core.Descriptor, err error) {
	// TODO: verify constraints (but not here)
	d.PrimaryIP = netip.Addr(s.P)
	d.WireGuardPK = s.PK
	d.CsidhPK = s.C
	d.Addrs = nil
	for _, addr := range s.V4A {
		d.Addrs = append(d.Addrs, netip.Addr(addr))
	}
	for _, addr := range s.V6A {
		d.Addrs = append(d.Addrs, netip.Addr(addr))
	}
	d.VulaPK = s.VK
	d.ValidStart = s.VF
	d.ValidDuration = s.DT
	port := s.Port
	if port < 0 || port > math.MaxUint16 {
		err = fmt.Errorf("port %d out of range", port)
		return
	}
	d.WireGuarPort = uint16(port)
	d.Hostname = s.Hostname
	d.Routes = make([]netip.Prefix, len(s.R))
	for i := range d.Routes {
		d.Routes[i] = netip.Prefix(s.R[i])
	}
	d.Ephemeral = s.E
	d.Signature = s.S
	return
}

func ExportDescriptor(d *core.Descriptor) (s Descriptor) {
	s.P = IPAddr(d.PrimaryIP)
	s.PK = d.WireGuardPK
	s.C = d.CsidhPK
	for _, addr := range d.Addrs {
		switch {
		case addr.Is4():
			s.V4A = append(s.V4A, IPAddr(addr))
		case addr.Is6():
			s.V6A = append(s.V6A, IPAddr(addr))
		}
	}
	s.VK = d.VulaPK
	s.VF = d.ValidStart
	s.DT = d.ValidDuration
	s.Port = int64(d.WireGuarPort)
	s.Hostname = d.Hostname
	s.R = make(IPPrefixList, len(d.Routes))
	for i := range s.R {
		s.R[i] = IPPrefix(d.Routes[i])
	}
	s.E = d.Ephemeral
	s.S = d.Signature
	return s
}

func ImportSystemState(s *SystemState) (p core.SystemState) {
	p.CurrentSubnets = make(map[netip.Prefix][]netip.Addr, len(s.CurrentSubnets))
	for prefix, addrs := range s.CurrentSubnets {
		addrsConverted := make([]netip.Addr, len(addrs))
		for i, addr := range addrs {
			addrsConverted[i] = netip.Addr(addr)
		}
		p.CurrentSubnets[netip.Prefix(prefix)] = addrsConverted
	}
	p.CurrentInterfaces = make(map[string][]netip.Addr, len(s.CurrentInterfaces))
	for name, addrs := range s.CurrentInterfaces {
		addrsConverted := make([]netip.Addr, len(addrs))
		for i, addr := range addrs {
			addrsConverted[i] = netip.Addr(addr)
		}
		p.CurrentInterfaces[name] = addrsConverted
	}
	p.OurWgPK = s.OurWgPK
	p.Gateways = make([]netip.Addr, len(s.Gateways))
	for i := range p.Gateways {
		p.Gateways[i] = netip.Addr(s.Gateways[i])
	}
	p.HasV6 = s.HasV6
	return
}

func ExportSystemState(p *core.SystemState) (s SystemState) {
	s.CurrentSubnets = make(Map[IPPrefix, Slice[IPAddr]], len(p.CurrentSubnets))
	for prefix, addrs := range p.CurrentSubnets {
		addrsConverted := make([]IPAddr, len(addrs))
		for i, addr := range addrs {
			addrsConverted[i] = IPAddr(addr)
		}
		s.CurrentSubnets[IPPrefix(prefix)] = addrsConverted
	}
	s.CurrentInterfaces = make(Map[string, Slice[IPAddr]], len(p.CurrentInterfaces))
	for name, addrs := range p.CurrentInterfaces {
		addrsConverted := make([]IPAddr, len(addrs))
		for i, addr := range addrs {
			addrsConverted[i] = IPAddr(addr)
		}
		s.CurrentInterfaces[name] = addrsConverted
	}
	s.OurWgPK = p.OurWgPK
	s.Gateways = make(Slice[IPAddr], len(p.Gateways))
	for i := range s.Gateways {
		s.Gateways[i] = IPAddr(p.Gateways[i])
	}
	s.HasV6 = p.HasV6
	return
}

func ImportPeer(s *Peer) (p core.Peer, err error) {
	p.Descriptor, err = ImportDescriptor(&s.Descriptor)
	if err != nil {
		return
	}
	p.Petname = s.Petname
	p.Nicknames = make(map[string]bool, len(s.Nicknames))
	for k, v := range s.Nicknames {
		p.Nicknames[k] = v
	}
	p.Addrs = make(map[netip.Addr]bool)
	for k, v := range s.IPv4Addrs {
		p.Addrs[netip.Addr(k)] = v
	}
	for k, v := range s.IPv6Addrs {
		p.Addrs[netip.Addr(k)] = v
	}
	p.Enabled = s.Enabled
	p.Verified = s.Verified
	p.Pinned = s.Pinned
	p.UseAsGateway = s.UseAsGateway
	return
}

func ExportPeer(p *core.Peer) (s Peer) {
	s.Descriptor = ExportDescriptor(&p.Descriptor)
	s.Petname = p.Petname
	s.Nicknames = make(Map[string, bool])
	for k, v := range p.Nicknames {
		s.Nicknames[k] = v
	}
	s.IPv4Addrs = make(Map[IPAddr, bool])
	s.IPv6Addrs = make(Map[IPAddr, bool])
	for k, v := range p.Addrs {
		if k.Is4() {
			s.IPv4Addrs[IPAddr(k)] = v
		} else if k.Is6() {
			s.IPv6Addrs[IPAddr(k)] = v
		}
	}
	s.Enabled = p.Enabled
	s.Verified = p.Verified
	s.Pinned = p.Pinned
	s.UseAsGateway = p.UseAsGateway
	return
}

func ImportPrefs(s *Prefs) (p core.Prefs) {
	p.PinNewPeers = s.PinNewPeers
	p.AutoRepair = s.AutoRepair
	p.SubnetsAllowed = make([]netip.Prefix, len(s.SubnetsAllowed))
	for i := range p.SubnetsAllowed {
		p.SubnetsAllowed[i] = netip.Prefix(s.SubnetsAllowed[i])
	}
	p.SubnetsForbidden = make([]netip.Prefix, len(s.SubnetsForbidden))
	for i := range p.SubnetsForbidden {
		p.SubnetsForbidden[i] = netip.Prefix(s.SubnetsForbidden[i])
	}
	p.IfacePrefixAllowed = s.IfacePrefixAllowed
	p.AcceptNonlocal = s.AcceptNonlocal
	p.LocalDomains = s.LocalDomains
	p.EphemeralMode = s.EphemeralMode
	p.AcceptDefaultRoute = s.AcceptDefaultRoute
	p.OverwriteUnpinned = s.OverwriteUnpinned
	p.ExpireTime = time.Second * time.Duration(s.ExpireTime)
	p.PrimaryIP = netip.Addr(s.PrimaryIP)
	p.RecordEvents = s.RecordEvents
	p.EnableIPv6 = s.EnableIPv6
	p.EnableIPv4 = s.EnableIPv4
	return
}

func ExportPrefs(p *core.Prefs) (s Prefs) {
	s.PinNewPeers = p.PinNewPeers
	s.AutoRepair = p.AutoRepair
	s.SubnetsAllowed = make(Slice[IPPrefix], len(p.SubnetsAllowed))
	for i := range s.SubnetsAllowed {
		s.SubnetsAllowed[i] = IPPrefix(p.SubnetsAllowed[i])
	}
	s.SubnetsForbidden = make(Slice[IPPrefix], len(p.SubnetsForbidden))
	for i := range s.SubnetsForbidden {
		s.SubnetsForbidden[i] = IPPrefix(p.SubnetsForbidden[i])
	}
	s.IfacePrefixAllowed = p.IfacePrefixAllowed
	s.AcceptNonlocal = p.AcceptNonlocal
	s.LocalDomains = p.LocalDomains
	s.EphemeralMode = p.EphemeralMode
	s.AcceptDefaultRoute = p.AcceptDefaultRoute
	s.OverwriteUnpinned = p.OverwriteUnpinned
	s.ExpireTime = int64(p.ExpireTime.Seconds())
	s.PrimaryIP = IPAddr(p.PrimaryIP)
	s.RecordEvents = p.RecordEvents
	s.EnableIPv6 = p.EnableIPv6
	s.EnableIPv4 = p.EnableIPv4
	return
}

func ImportState(s *OrganizeState) (p core.OrganizeState, err error) {
	prefs := ImportPrefs(&s.Prefs)
	p.Prefs = &prefs
	p.Peers = make(map[string]*core.Peer)
	for id, peer := range s.Peers {
		peer, err := ImportPeer(&peer)
		if err != nil {
			return p, err
		}
		p.Peers[id] = &peer
	}
	systemState := ImportSystemState(&s.SystemState)
	p.SystemState = &systemState
	p.EventLog = s.EventLog
	return
}

func ExportState(p *core.OrganizeState) (s OrganizeState) {
	s.Prefs = ExportPrefs(p.Prefs)
	s.Peers = make(Map[string, Peer])
	for id, peer := range p.Peers {
		s.Peers[id] = ExportPeer(peer)
	}
	s.SystemState = ExportSystemState(p.SystemState)
	s.EventLog = p.EventLog
	return
}
