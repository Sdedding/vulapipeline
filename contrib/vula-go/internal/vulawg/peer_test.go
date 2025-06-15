package vulawg

import (
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestPeerConfigFromWgCtrlV4(t *testing.T) {
	peer := createNetlinkPeer("192.168.0.0")

	c := peerConfigFromWgCtrl(peer)

	if addr := c.EndpointAddr.String(); addr != "192.168.0.0" {
		t.Errorf("c.EndpointAddr = '%s'", addr)
	}

	checkNetlinkPeerConfig(t, &c)
}

func TestPeerConfigFromWgCtrlV6(t *testing.T) {
	peer := createNetlinkPeer("FE80::FFFF:FFFF:FFFF:FFFE")

	c := peerConfigFromWgCtrl(peer)

	if addr := c.EndpointAddr.String(); strings.ToUpper(addr) != "FE80::FFFF:FFFF:FFFF:FFFE" {
		t.Errorf("c.EndpointAddr = '%s'", addr)
	}

	checkNetlinkPeerConfig(t, &c)
}

func createNetlinkPeer(addr string) *wgtypes.Peer {
	dummyKey := [wgtypes.KeyLen]byte{}
	for i := range dummyKey {
		dummyKey[i] = 'A'
	}

	ip, err := netip.ParseAddr(addr)
	if err != nil {
		panic(err)
	}

	return &wgtypes.Peer{
		PublicKey:                   dummyKey,
		PresharedKey:                dummyKey,
		PersistentKeepaliveInterval: 666,
		Endpoint: &net.UDPAddr{
			IP:   ip.AsSlice(),
			Port: 1000,
		},
		LastHandshakeTime: time.Unix(567, 0),
		ProtocolVersion:   99,
		TransmitBytes:     3,
		ReceiveBytes:      2,
	}
}

// checkNetlinkPeerConfig checks if the values in p match those
// hardcoded in createNetlinkPeer
func checkNetlinkPeerConfig(t *testing.T, c *core.PeerConfig) {
	if c.PersistentKeepalive != 666 {
		t.Errorf("c.PersistentKeepalive = %d", c.PersistentKeepalive)
	}

	if c.EndpointPort != 1000 {
		t.Errorf("c.EndpointPort = %d", c.EndpointPort)
	}

	if c.LatestHandshake.Unix() != 567 {
		t.Errorf("c.LatestHandshake = %d", c.LatestHandshake.Unix())
	}
}
