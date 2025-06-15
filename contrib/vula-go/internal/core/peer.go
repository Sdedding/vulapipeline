package core

import (
	"net/netip"
	"time"
)

type Peer struct {
	Descriptor   Descriptor
	Petname      string
	Nicknames    map[string]bool
	Addrs        map[netip.Addr]bool
	Enabled      bool
	Verified     bool
	Pinned       bool
	UseAsGateway bool
}

type PeerStats struct {
	LatestHandshake    int64
	RxBytes            int64
	TxBytes            int64
	HasLatestHandshake bool
}

type PeerConfig struct {
	Unspec              any
	Remove              bool
	PublicKey           []byte
	PresharedKey        []byte
	EndpointAddr        netip.Addr
	EndpointPort        uint16
	PersistentKeepalive time.Duration
	AllowedIPs          []netip.Prefix
	Stats               struct {
		RxBytes int64
		TxBytes int64
	}
	LatestHandshake time.Time
}
