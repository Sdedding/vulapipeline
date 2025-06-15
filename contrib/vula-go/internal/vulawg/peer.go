package vulawg

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"net/netip"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// applyPeerConfig syncs peer's wg config and routes. Returns a string
func applyPeerConfig(deviceName string, c *core.PeerConfig, dryRun bool) (log []string, err error) {
	client, err := wgctrl.New()
	if err != nil {
		return
	}
	defer util.CloseLog(client)

	device, err := client.Device(deviceName)
	if err != nil {
		return
	}

	publicKey, err := wgtypes.NewKey(c.PublicKey)
	if err != nil {
		return
	}

	var currentPeer *wgtypes.Peer
	for i := range device.Peers {
		if bytes.Equal(device.Peers[i].PublicKey[:], publicKey[:]) {
			currentPeer = &device.Peers[i]
			break
		}
	}

	if currentPeer != nil && c.Remove {
		// device exists and should be removed
		log = append(log, fmt.Sprintf("# removing wireguard peer %s", toBase64(c.PublicKey)))
		log = append(log, peerConfigToLog(deviceName, c))
		if !dryRun {
			err = client.ConfigureDevice(deviceName, wgtypes.Config{Peers: []wgtypes.PeerConfig{{PublicKey: publicKey, Remove: true}}})
		}
		return
	}
	if c.Remove {
		// device does not axist and should be removed
		log = append(log, fmt.Sprintf("# cant remove non-existent wireguard peer %s", c.PublicKey))
		log = append(log, peerConfigToLog(deviceName, c))
		return
	}

	// if we get here, the device should be created or updated if it exists
	// first we have to bring the data into the right format
	// and then create or update the device as needed

	presharedKey, err := wgtypes.NewKey(c.PresharedKey)
	if err != nil {
		return
	}
	var endpointAddr net.UDPAddr
	endpointAddr.IP = net.IP(c.EndpointAddr.AsSlice())
	endpointAddr.Port = int(c.EndpointPort)

	var allowedIPs []net.IPNet
	if c.AllowedIPs != nil {
		allowedIPs = make([]net.IPNet, len(c.AllowedIPs))
		for i, prefix := range c.AllowedIPs {
			allowedIPs[i].IP = net.IP(prefix.Addr().AsSlice())
			allowedIPs[i].Mask = net.CIDRMask(prefix.Bits(), prefix.Addr().BitLen())
		}
	}

	if currentPeer == nil {
		// device does not exist and should be created
		peerConfig := wgtypes.PeerConfig{
			PublicKey:                   publicKey,
			PresharedKey:                &presharedKey,
			Endpoint:                    &endpointAddr,
			PersistentKeepaliveInterval: &c.PersistentKeepalive,
			AllowedIPs:                  allowedIPs,
			ReplaceAllowedIPs:           true,
		}
		if !dryRun {
			err = client.ConfigureDevice(deviceName, wgtypes.Config{Peers: []wgtypes.PeerConfig{peerConfig}})
		}
		log = append(log, fmt.Sprintf("# configure new wireguard peer %s", toBase64(c.PublicKey)))
		log = append(log, peerConfigToLog(deviceName, c))
		return
	}

	// device exists and should be created
	// update configuration as needed
	peerConfig := wgtypes.PeerConfig{PublicKey: publicKey, UpdateOnly: true}

	log = append(log, fmt.Sprintf("# reconfigure wireguard peer %s", toBase64(c.PublicKey)))
	log = append(log, peerConfigToLog(deviceName, c))

	hasChange := false

	if !bytes.Equal(currentPeer.PresharedKey[:], c.PresharedKey) {
		peerConfig.PresharedKey = &presharedKey
		hasChange = true
	}

	if !currentPeer.Endpoint.IP.Equal(endpointAddr.IP) || currentPeer.Endpoint.Port != endpointAddr.Port {
		peerConfig.Endpoint = &endpointAddr
		hasChange = true
	}

	if currentPeer.PersistentKeepaliveInterval != c.PersistentKeepalive {
		peerConfig.PersistentKeepaliveInterval = &c.PersistentKeepalive
		hasChange = true
	}

checkAllowedIPs:
	for _, p := range allowedIPs {
		for _, currentNet := range currentPeer.AllowedIPs {
			if !p.IP.Equal(currentNet.IP) || !bytes.Equal(p.Mask, currentNet.Mask) {
				peerConfig.AllowedIPs = allowedIPs
				peerConfig.ReplaceAllowedIPs = true
				hasChange = true
				break checkAllowedIPs
			}
		}
	}

	if !hasChange {
		core.LogDebug("apply_peerconfig: no wg update necessary")
		return
	}

	if !dryRun {
		err = client.ConfigureDevice(deviceName, wgtypes.Config{
			Peers: []wgtypes.PeerConfig{peerConfig},
		})
	}
	return
}

func getPeers(deviceName string) (peers []core.PeerConfig, err error) {
	client, err := wgctrl.New()
	if err != nil {
		return
	}
	defer util.CloseLog(client)

	device, err := client.Device(deviceName)
	if err != nil {
		return
	}

	peers = make([]core.PeerConfig, len(device.Peers))
	for i := range peers {
		peers[i] = peerConfigFromWgCtrl(&device.Peers[i])
	}
	return
}

// peerconfigFromNetlink converts approximately from what pyroute2 produces to what pyroute2 consumes
func peerConfigFromWgCtrl(peer *wgtypes.Peer) core.PeerConfig {
	allowedIPs := make([]netip.Prefix, len(peer.AllowedIPs))
	for i := range peer.AllowedIPs {
		var err error
		allowedIPs[i], err = ipNetToPrefix(&peer.AllowedIPs[i])
		if err != nil {
			panic(err)
		}
	}

	var endpointAddr netip.Addr
	var endpointPort uint16

	if peer.Endpoint != nil {
		var ok bool
		endpointAddr, ok = netip.AddrFromSlice(peer.Endpoint.IP)
		if !ok {
			panic("vula: received invalid wg peer addr from wgctrl")
		}

		port := peer.Endpoint.Port
		if port < 0 || port > math.MaxUint16 {
			panic("vula: received invalid wg endpoint port from wgctrl")
		}
		endpointPort = uint16(port)
	}

	c := core.PeerConfig{
		PublicKey:           peer.PublicKey[:],
		PresharedKey:        peer.PresharedKey[:],
		EndpointAddr:        endpointAddr,
		EndpointPort:        endpointPort,
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: peer.PersistentKeepaliveInterval,
		LatestHandshake:     peer.LastHandshakeTime,
	}

	c.Stats.RxBytes = peer.ReceiveBytes
	c.Stats.TxBytes = peer.TransmitBytes

	return c
}

func peerConfigToLog(deviceName string, c *core.PeerConfig) string {
	// complicated string formatting
	remove := ""
	if c.Remove {
		remove = "remove "
	}

	endpoint := ""
	if !c.EndpointAddr.IsValid() {
		endpoint = fmt.Sprintf("endpoint %s:%d", c.EndpointAddr, c.EndpointPort)
	}

	argsBuilder := []byte{}
	if c.PersistentKeepalive != 0 {
		argsBuilder = append(argsBuilder, fmt.Sprintf("persistent_keepalive %s ", c.PersistentKeepalive)...)
	}
	if c.PresharedKey != nil {
		argsBuilder = append(argsBuilder, fmt.Sprintf("preshared_key %s ", toBase64(c.PresharedKey))...)
	}
	args := string(argsBuilder)

	allowedIPs := ""
	if c.AllowedIPs != nil {
		l := make([]string, len(c.AllowedIPs))
		for i, ip := range c.AllowedIPs {
			l[i] = ip.String()
		}
		allowedIPs = fmt.Sprintf("allowed-ips %s", strings.Join(l, ", "))
	}

	setComand := fmt.Sprintf("vula wg set %s peer %s %s%s%s%s",
		deviceName, toBase64(c.PublicKey), remove, endpoint, args, allowedIPs)

	return setComand
}
