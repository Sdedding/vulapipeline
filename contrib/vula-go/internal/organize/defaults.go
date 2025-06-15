package organize

import (
	"net/netip"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

var defaultSystemState = &core.SystemState{
	OurWgPK:        make([]byte, 32),
	CurrentSubnets: make(map[netip.Prefix][]netip.Addr),
	Gateways:       []netip.Addr{},
}

var defaultOrganizeState = &core.OrganizeState{
	Prefs:       defaultPrefs,
	Peers:       make(map[string]*core.Peer),
	SystemState: defaultSystemState,
	EventLog:    []string{},
}

var defaultPrefs = &core.Prefs{
	PinNewPeers:    false,
	AcceptNonlocal: false,
	AutoRepair:     true,
	SubnetsAllowed: []netip.Prefix{
		ipv6LL,
		ipv6ULA,
		ipv4LL,
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("172.16.0.0/12"),
	},
	SubnetsForbidden: []netip.Prefix{},
	IfacePrefixAllowed: []string{
		"en",
		"eth",
		"wl",
		"thunderbolt",
	},
	LocalDomains: []string{
		"local.",
		"local",
	},
	EphemeralMode:      false,
	AcceptDefaultRoute: true,
	RecordEvents:       false,
	ExpireTime:         time.Second * 3600,
	OverwriteUnpinned:  true,
	PrimaryIP:          netip.Addr{},
	EnableIPv6:         true,
	EnableIPv4:         true,
}
