package organize

import (
	"bytes"
	"fmt"
	"net/netip"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

type systemStateConfig struct {
	WireguardPK        []byte
	IfacePrefixAllowed []string
	PrimaryIP          netip.Addr
	SubnetsForbidden   []netip.Prefix
	EnableIPv4         bool
	EnableIPv6         bool
}

// syncPeer syncs peer's wg config and routes. Returns a string
func (o *O) syncPeer(peers map[string]*core.Peer, id string, systemState *core.SystemState, routingTable int, dryRun bool) (log []string, err error) {
	p := peers[id]
	if p == nil {
		err = fmt.Errorf("peer %s", id)
		return
	}

	if p.Enabled {
		core.LogDebugf("Syncing enabled peer %s", peerName(p))

		ctidhPsk, err := o.dh.GetPSK(p.Descriptor.CsidhPK)
		if err != nil {
			return log, err
		}

		config := peerWireGuardConfig(p, ctidhPsk)
		applyLog, err := o.wgi.SetPeer(config, dryRun)
		log = append(log, applyLog...)
		if err != nil {
			return log, err
		}

		syncLog, err := o.config.NetworkSystem.SyncRoutes(peerRoutes(p), routingTable, o.wgi.Name(), systemState.CurrentSubnets, dryRun)
		log = append(log, syncLog...)
		if err != nil {
			return log, err
		}

		if p.UseAsGateway {
			syncLog, err = o.config.NetworkSystem.SyncRoutes(core.GatewayRoutes, core.LinuxMainRoutingTable, o.wgi.Name(), systemState.CurrentSubnets, dryRun)
			log = append(log, syncLog...)
			if err != nil {
				return log, err
			}
			core.LogDebugf("organize.Peer.sync result: %s", log)
		}
	} else {
		// FIXME: this should go away with triggers, but hasn't yet?
		removeLog, err := o.removeUnknown(routingTable, peers, dryRun)
		log = append(log, removeLog...)
		if err != nil {
			return log, err
		}
	}
	return
}

// removeUnknown This is currently the code path where disabled and removed peers get
// their routes and wg peer configs removed. In the future, deferred
// actions from the event engine should remove the specific things that we
// know need to be removed, and then this method will actually only be
// used to remove rogue entries.
func (o *O) removeUnknown(routingTable int, peers map[string]*core.Peer, dryRun bool) (log []string, err error) {
	enabledPeers := []*core.Peer{}
	for _, peer := range peers {
		if peer.Enabled {
			enabledPeers = append(enabledPeers, peer)
		}
	}

	enabledPKs := make(map[string]struct{}, len(enabledPeers))
	for _, peer := range enabledPeers {
		enabledPKs[toBase64String(peer.Descriptor.WireGuardPK)] = struct{}{}
	}

	wgiPeers, err := o.wgi.Peers()
	if err != nil {
		return
	}
	for _, peer := range wgiPeers {
		pk := toBase64String(peer.PublicKey)
		if _, ok := enabledPKs[pk]; !ok {
			var removeLog []string
			if !dryRun {
				removeLog, err = o.wgi.SetPeer(&core.PeerConfig{PublicKey: peer.PublicKey, Remove: true}, dryRun)
				log = append(log, removeLog...)
				if err != nil {
					return
				}
				core.LogInfof("Removing unexpected peer pk: %s", toBase64String(peer.PublicKey))
			}
			log = append(log, fmt.Sprintf("wg set %s peer %s remove", o.wgi.Name(), pk))
		}
	}

	expectedRoutes := map[netip.Prefix]struct{}{}
	for _, peer := range enabledPeers {
		for _, dst := range peerRoutes(peer) {
			expectedRoutes[dst] = struct{}{}
		}
	}
	hasEnabledGateway := false
	for _, p := range enabledPeers {
		if p.UseAsGateway {
			hasEnabledGateway = true
			break
		}
	}

	removeRouteLog, err := o.config.NetworkSystem.RemoveUnknownRoutes(expectedRoutes, routingTable, hasEnabledGateway, dryRun)
	log = append(log, removeRouteLog...)
	return
}

func (o *O) syncInterfaces(privateKey []byte, listenPort uint16, firewallMark int, primaryIP netip.Addr, dryRun bool) (log []string, err error) {
	log, err = o.config.WireguardSystem.SyncInterface(o.config.InterfaceName, primaryIP, dryRun)
	if err != nil {
		return
	}

	currentPrivateKey, _, currentListenPort, currentFirewallMark, err := o.wgi.Configuration()
	if err != nil {
		return
	}

	ok := bytes.Equal(privateKey, currentPrivateKey) &&
		listenPort == currentListenPort &&
		firewallMark == currentFirewallMark

	if ok {
		core.LogDebugf("No reconfiguration needed for interface %s", o.config.InterfaceName)
		return
	}

	if !dryRun {
		err = o.wgi.SetConfiguration(privateKey, listenPort, firewallMark)
		if err != nil {
			return
		}
	}
	log = append(log, "# configure interface")
	log = append(log, fmt.Sprintf("WireGuard.set(%s, privateKey = ****, listenPort = %d, firewallMark = %d)", o.config.InterfaceName, listenPort, firewallMark))
	return
}
