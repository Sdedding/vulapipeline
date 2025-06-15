package organize

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"
	"testing"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func TestSystemStateCurrentIPs(t *testing.T) {
	type testCase struct {
		subnets map[netip.Prefix][]netip.Addr
		wants   []netip.Addr
	}

	tests := []testCase{
		{
			map[netip.Prefix][]netip.Addr{
				netip.MustParsePrefix("10.0.0.0/24"): {
					netip.MustParseAddr("10.0.0.1"),
					netip.MustParseAddr("10.0.0.2"),
				},
			},
			[]netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.2"),
			},
		},
		{
			map[netip.Prefix][]netip.Addr{
				netip.MustParsePrefix("10.0.0.0/24"): {
					netip.MustParseAddr("10.0.0.1"),
				},
				netip.MustParsePrefix("192.168.0.0/24"): {
					netip.MustParseAddr("192.168.1.1"),
				},
			},
			[]netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("192.168.1.1"),
			},
		},
		{
			map[netip.Prefix][]netip.Addr{
				netip.MustParsePrefix("FE80::/10"): {
					netip.MustParseAddr("FE80::FFFF:FFFE"),
					netip.MustParseAddr("FE80::FFFF:FFFD"),
				},
			},
			[]netip.Addr{
				netip.MustParseAddr("fe80::ffff:fffe"),
				netip.MustParseAddr("fe80::ffff:fffd"),
			},
		},
		{
			map[netip.Prefix][]netip.Addr{
				netip.MustParsePrefix("FE80::/10"): {
					netip.MustParseAddr("FE80::FFFF:FFFE"),
				},
				netip.MustParsePrefix("FC00::/7"): {
					netip.MustParseAddr("FC00::FFFF:FFFE"),
				},
			},
			[]netip.Addr{
				netip.MustParseAddr("fe80::ffff:fffe"),
				netip.MustParseAddr("fc00::ffff:fffe"),
			},
		},
	}

	for _, test := range tests {
		s := core.SystemState{
			CurrentSubnets: test.subnets,
		}

		currentIPs := systemStateCurrentIPs(&s)
		slices.SortFunc(currentIPs, func(a, b netip.Addr) int {
			return strings.Compare(fmt.Sprint(b), fmt.Sprint(a))
		})
		slices.SortFunc(test.wants, func(a, b netip.Addr) int {
			return strings.Compare(fmt.Sprint(b), fmt.Sprint(a))
		})

		if len(currentIPs) != len(test.wants) {
			t.Errorf("len(currentIPs) = %d", len(currentIPs))
		}

		for i, ip := range currentIPs {
			if ip != test.wants[i] {
				t.Errorf("currentIPs[%d] = %s", i, ip)
			}
		}
	}
}
