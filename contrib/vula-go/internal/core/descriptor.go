package core

import (
	"net/netip"
)

type Descriptor struct {
	PrimaryIP     netip.Addr     // p
	WireGuardPK   []byte         // pk
	CsidhPK       []byte         // c
	Addrs         []netip.Addr   // v4a, v6a
	VulaPK        []byte         // vk
	ValidStart    int64          // vf
	ValidDuration int64          // dt
	WireGuarPort  uint16         // port
	Hostname      string         // hostname
	Routes        []netip.Prefix // r
	Ephemeral     bool           // e
	Signature     []byte         // s
}
