package core

import "net/netip"

// LinuxMainRoutingTable specifies the main routing table on linux
// This should be abstracted if we want to support something different than linux
const LinuxMainRoutingTable = 254

var GatewayRoutes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/1"),
	netip.MustParsePrefix("128.0.0.0/1"),
	netip.MustParsePrefix("::/1"),
	netip.MustParsePrefix("8000::/1"),
}

var VulaSubnet = netip.MustParsePrefix("fdff:ffff:ffdf::/48")

const DummyLinkName = "vula-net"

const HostsFileName = "/var/lib/vula-organize/hosts"

const VulaOrganizeLibDir = "/var/lib/vula-organize"
