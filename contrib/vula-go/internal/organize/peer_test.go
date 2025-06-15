package organize

import (
	"net/netip"
	"strings"
	"testing"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

func TestPeerNamePetname(t *testing.T) {

	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "george",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	name := peerName(p)
	if name != "george" {
		t.Errorf("name = %s", name)
	}
}

func TestPeerNameHostname(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	name := peerName(p)
	if name != "george.local" {
		t.Errorf("name = %s", name)
	}
}

func TestPeerNameNickname(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: false,
			"schnubbi": true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	name := peerName(p)
	if name != "schnubbi" {
		t.Errorf("name = %s", name)
	}
}

func TestOtherNamesEmpty(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestOtherNamesPetname(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "george",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "george.local" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestOtherNamesPetnameNickname(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "george",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
			"schnubbi": true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "george.local, schnubbi" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestOtherNamesEmptyV6(t *testing.T) {
	d := getTestPeerDescriptorIPv6(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestOtherNamesPetnameV6(t *testing.T) {
	d := getTestPeerDescriptorIPv6(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "george",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "george.local" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestOtherNamesPetnameNicknameV6(t *testing.T) {
	d := getTestPeerDescriptorIPv6(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "george",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
			"schnubbi": true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	otherNames := strings.Join(peerOtherNames(p), ", ")
	if otherNames != "george.local, schnubbi" {
		t.Errorf("otherNames = %s", otherNames)
	}
}

func TestNameAndIdV4(t *testing.T) {
	d := getTestPeerDescriptorIPv4(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	nameAndID := peerNameAndID(p)
	if nameAndID != "george.local (Q0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0M=)" {
		t.Errorf("nameAndID = %s", nameAndID)
	}
}

func TestNameAndIdV6(t *testing.T) {
	d := getTestPeerDescriptorIPv6(t)

	p := &core.Peer{
		Descriptor: d,
		Petname:    "",
		Pinned:     false,
		Enabled:    true,
		Verified:   false,
		Nicknames: map[string]bool{
			d.Hostname: true,
		},
		Addrs: enabledIpAddrs(d.Addrs, true),
	}

	nameAndID := peerNameAndID(p)
	if nameAndID != "george.local (Q0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0NDQ0M=)" {
		t.Errorf("nameAndID = %s", nameAndID)
	}
}

func enabledIpAddrs(addrs []netip.Addr, enabled bool) map[netip.Addr]bool {
	m := map[netip.Addr]bool{}
	for _, addr := range addrs {
		m[addr] = enabled
	}
	return m
}

func getTestPeerDescriptorIPv4(t *testing.T) core.Descriptor {
	input := "v4a=192.168.6.9;c=QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUF" +
		"BQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQQ==;dt=86400;e=0;hostname=george.local;pk=QkJCQkJCQ" +
		"kJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkI=;port=123;vf=1601388653;vk=Q0NDQ0NDQ0NDQ0NDQ0NDQ0N" +
		"DQ0NDQ0NDQ0NDQ0NDQ0M="

	s, err := schema.ParseDescriptorString(input)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := schema.ImportDescriptor(&s)
	return d
}

func getTestPeerDescriptorIPv6(t *testing.T) core.Descriptor {
	input := "v6a=fe80::377b:d17:9b74:1b91;c=QUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUF" +
		"BQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQQ==;dt=86400;e=0;hostname=george.local;pk=QkJCQkJCQ" +
		"kJCQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkI=;port=123;vf=1601388653;vk=Q0NDQ0NDQ0NDQ0NDQ0NDQ0N" +
		"DQ0NDQ0NDQ0NDQ0NDQ0M="

	s, err := schema.ParseDescriptorString(input)
	if err != nil {
		t.Fatal(err)
	}

	d, _ := schema.ImportDescriptor(&s)
	return d
}
