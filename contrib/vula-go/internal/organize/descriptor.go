package organize

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"os"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

func descriptorToString(d *core.Descriptor) string {
	s := schema.ExportDescriptor(d)
	return schema.SerializeDescriptorString(&s)
}

func descriptorID(d *core.Descriptor) string {
	return base64.StdEncoding.EncodeToString(d.VulaPK)
}

func makePeerFromDescriptor(d *core.Descriptor) *core.Peer {
	addrs := make(map[netip.Addr]bool, len(d.Addrs))
	for _, addr := range d.Addrs {
		addrs[addr] = true
	}

	return &core.Peer{
		Descriptor: *d,
		Enabled:    true,
		Nicknames:  map[string]bool{d.Hostname: true},
		Addrs:      addrs,
	}
}

func checkDescriptorFreshness(d *core.Descriptor, currentTime time.Time) bool {
	now := currentTime.UTC().Unix()
	validEnd := now + d.ValidDuration
	if validEnd < now {
		return false // integer overflow
	}
	return d.ValidStart >= now && validEnd > now
}

func constructServiceDescriptor(keys *core.Keys, state *core.OrganizeState, port uint16, ips []netip.Addr, validFrom int64) core.Descriptor {
	core.LogInfof("Constructing service descriptor id: %d", validFrom) // hö

	addrs := []netip.Addr{}
	switch {
	case state.Prefs.EnableIPv4 && state.Prefs.EnableIPv6:
		addrs = ips
	case state.Prefs.EnableIPv4:
		for _, addr := range ips {
			if addr.Is4() {
				addrs = append(addrs, addr)
			}
		}
	case state.Prefs.EnableIPv6:
		for _, addr := range ips {
			if addr.Is6() {
				addrs = append(addrs, addr)
			}
		}
	}

	sortLLFirst(addrs)

	d := core.Descriptor{
		PrimaryIP:     state.Prefs.PrimaryIP,
		WireGuardPK:   keys.WgCurve25519.PK,
		CsidhPK:       keys.PqCtidhP512.PK,
		Addrs:         addrs,
		VulaPK:        keys.VkEd25519.PK,
		ValidStart:    validFrom,
		ValidDuration: 86400,
		WireGuarPort:  uint16(port),
		Hostname:      getHostname(),
		Routes:        nil,
		Ephemeral:     false,
	}
	return d
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	//TODO: HACK: this is probably not right (but python does the same)
	// FIXME: make this a pref
	return fmt.Sprintf("%s.local.", hostname)
}
