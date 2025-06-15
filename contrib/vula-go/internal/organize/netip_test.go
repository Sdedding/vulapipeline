package organize

import (
	"net/netip"
	"testing"
)

func TestSortLLFirst(t *testing.T) {
	// arrange
	addrs := []netip.Addr{
		netip.MustParseAddr("169.254.0.1"), netip.MustParseAddr("127.0.0.1"),
		netip.MustParseAddr("ff00::1"), netip.MustParseAddr("169.254.0.2"),
		netip.MustParseAddr("fe80::1"), netip.MustParseAddr("0::1"),
	}

	// act
	sortLLFirst(addrs)

	// assert
	expectedAddrs := []string{"fe80::1", "169.254.0.1",
		"169.254.0.2", "ff00::1", "::1",
		"127.0.0.1"}

	for i := range addrs {
		if addrs[i].String() != expectedAddrs[i] {
			t.Errorf("addrs[%d] = %s", i, addrs[i])
		}
	}
}
