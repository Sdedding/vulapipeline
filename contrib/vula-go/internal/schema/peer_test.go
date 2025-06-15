package schema

import (
	"net/netip"
	"testing"
)

func TestPeerRead(t *testing.T) {
	// arrange
	p := Peer{}
	p.Descriptor.Hostname = "test"
	p.Enabled = true
	p.Descriptor.V4A = IPAddrList{IPAddr(netip.MustParseAddr("192.168.0.1"))}

	// act
	hostname := ""
	err := Read(p, []string{"descriptor", "hostname"}, &hostname)
	if err != nil {
		t.Error(err)
	}

	enabled := false
	err = Read(p, []string{"enabled"}, &enabled)
	if err != nil {
		t.Error(err)
	}

	v4a := []netip.Addr{}
	err = Read(&p, []string{"descriptor", "v4a"}, &v4a)
	if err != nil {
		t.Error(err)
	}

	// assert
	if hostname != "test" {
		t.Errorf("hostname = '%s'", hostname)
	}
	if !enabled {
		t.Errorf("enabled = %v", enabled)
	}
	if len(v4a) != 1 || v4a[0].String() != "192.168.0.1" {
		t.Errorf("v4a = %v", v4a)
	}
}
