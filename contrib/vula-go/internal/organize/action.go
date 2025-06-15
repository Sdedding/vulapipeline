package organize

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

// peerAddrAdd adds an ip address to the specified peer
func (e *organizeTransaction) peerAddrAdd(vk string, ip netip.Addr) error {
	e.AddAction("PEER_ADDR_ADD", vk, ip.String())

	peer := e.State.Peers[vk]
	if peer == nil {
		return &ErrPeerNotFound{"vk", vk}
	}

	peer.Addrs[ip] = true
	e.AddTrigger(&syncPeerTrigger{vk})
	return nil
}

func (e *organizeTransaction) peerAddrDel(vk string, ip netip.Addr) error {
	e.AddAction("PEER_ADDR_DEL", vk, ip.String())

	peer := e.State.Peers[vk]
	if peer == nil {
		return &ErrPeerNotFound{"vk", vk}
	}

	delete(peer.Addrs, ip)

	e.AddTrigger(&syncPeerTrigger{vk})
	prefix, _ := ip.Prefix(ip.BitLen())
	e.AddTrigger(&removeRoutesTrigger{[]netip.Prefix{prefix}})

	return nil
}

func (e *organizeTransaction) edit(operation editOperation, path []string, value any) error {
	e.AddAction("EDIT", strings.Join(path, ", "), fmt.Sprintf("%v", value))
	s := schema.ExportState(e.State)
	var err error
	switch operation {
	case editAdd:
		err = schema.Add(value, path, &s)
	case editRemove:
		err = schema.Remove(value, path, &s)
	case editSet:
		err = schema.Write(value, path, &s)
	}
	if err != nil {
		return err
	}

	*e.State, err = schema.ImportState(&s)
	if err != nil {
		return err
	}
	if len(path) > 1 {
		if path[0] == "peers" {
			err = e.updatePeer(path[1], nil, nil)
		}
	} else if len(path) > 0 && path[0] == "prefs" {
		e.AddTrigger(&getNewSystemStateTrigger{})
	}
	e.AddTrigger(&removeUnknownTrigger{})
	return err
}

func (e *organizeTransaction) adjustToNewSystemState(newSystemState *core.SystemState) error {
	e.AddAction("ADJUST_TO_NEW_SYSTEM_STATE", fmt.Sprintf("%v", newSystemState))

	gateways := peersGateways(e.State.Peers)
	if len(gateways) > 0 &&
		(!gateways[0].Pinned) &&
		(len(getOverlap(peerEnabledIPs(gateways[0], true), newSystemState.Gateways)) == 0) {
		// if a non-pinned peer had our gateway IP but no longer does,
		// remove its gateway flag

		gateways[0].UseAsGateway = false
		e.AddTrigger(&removeRoutesTrigger{core.GatewayRoutes})
	}

	if !(len(gateways) > 0 && gateways[0].Pinned) {
		// if there isn't a pinned peer acting as the gateway.
		// FIXME: this could set two peers as the gateway if the system has
		// multiple default routes! we need to think about how to handle
		// that...

		for _, gateway := range newSystemState.Gateways {
			hit, ok := peersGetWithEnabledIP(e.State.Peers, gateway)
			if ok {
				hit.UseAsGateway = true
				e.AddTrigger(&syncPeerTrigger{peerID(hit)})

				// first hit wins (there could be multiples if we have
				// multiple default routes, but only one can get
				// allowedips=/0 so we just take the first one.)
				break
			}
		}
	}

	for id := range e.State.Peers {
		err := e.updatePeer(id, nil, newSystemState)
		if err != nil {
			return err
		}
	}

	e.State.SystemState = newSystemState

	//TODO: remove endpoints from pinned peers that became non-local
	return nil
}

func (e *organizeTransaction) acceptNewPeer(d *core.Descriptor) error {
	s := schema.ExportDescriptor(d)
	e.AddAction("ACCEPT_NEW_PEER", schema.SerializeDescriptorString(&s))
	var err error

	peer := makePeerFromDescriptor(d)
	if d.Ephemeral {
		peer.Pinned = false
	} else {
		peer.Pinned = e.State.Prefs.PinNewPeers
	}

	peerID := descriptorID(d)
	e.State.Peers[peerID] = peer
	err = e.updatePeer(peerID, nil, nil)
	return err
}

func (e *organizeTransaction) updatePeerDescriptor(peerID string, d *core.Descriptor) error {
	s := schema.ExportDescriptor(d)
	e.AddAction("UPDATE_PEER_DESCRIPTOR", peerID, schema.SerializeDescriptorString(&s))

	peer := e.State.Peers[peerID]
	if peer == nil {
		return &ErrPeerNotFound{"vk", peerID}
	}
	peer.Descriptor = *d
	return e.updatePeer(peerID, d, nil)
}

func (e *organizeTransaction) removePeer(peerID string) error {
	e.AddAction("REMOVE_PEER", peerID)

	peer := e.State.Peers[peerID]
	if peer == nil {
		return &ErrPeerNotFound{"vk", peerID}
	}

	delete(e.State.Peers, peerID)

	e.AddTrigger(&removeWgPeerTrigger{peer.Descriptor.WireGuardPK})
	e.AddTrigger(&removeRoutesTrigger{peerRoutes(peer)})

	if peer.UseAsGateway {
		// FIXME: this doesn't specify the routing table. maybe need to make
		// triggers api accept kwargs too?
		e.AddTrigger(&removeRoutesTrigger{core.GatewayRoutes})

		// these routes are currently only removed because we still call
		// sync (aka full repair) on system state change
	}
	return nil
}

func (e *organizeTransaction) updatePeer(peerID string, desc *core.Descriptor, systemState *core.SystemState) error {
	e.AddAction("UPDATE_PEER")

	peer := e.State.Peers[peerID]
	if peer == nil {
		return &ErrPeerNotFound{"vk", peerID}
	}
	core.LogInfof("calling updatePeer for %s", peerNameAndID(peer))
	if desc == nil {
		desc = &peer.Descriptor
	}
	if systemState == nil {
		systemState = e.State.SystemState
	}

	var subnets []netip.Prefix
	currentSubnetsNoULA := systemStateCurrentSubnetsNoULA(systemState)
	for prefix := range currentSubnetsNoULA {
		if prefix.IsValid() {
			subnets = append(subnets, prefix)
		}
	}
	allowedSubnets := e.State.Prefs.SubnetsAllowed
	newPeerAddrs := make(map[netip.Addr]bool)

	if peer.Pinned {
		for _, addr := range desc.Addrs {
			if addr.IsValid() {
				newPeerAddrs[addr] = true
			}
		}
		for addr, v := range peer.Addrs {
			if addr.IsValid() && v {
				newPeerAddrs[addr] = true
			}
		}
	} else {
		for _, ip := range desc.Addrs {
			if !ip.IsValid() {
				continue
			}
			for _, prefix := range append(subnets, allowedSubnets...) {
				if prefix.IsValid() && prefix.Contains(ip) {
					newPeerAddrs[ip] = true
				}
			}
		}
	}
	peer.Addrs = newPeerAddrs

	hasActiveIP := false
	for _, v := range peer.Addrs {
		if v {
			hasActiveIP = true
			break
		}
	}

	if !hasActiveIP {
		core.LogInfof("removing %s because it has no currently local IPs", peerNameAndID(peer))
		err := e.removePeer(peerID)
		if err != nil {
			return err
		}
	}

	if desc.Hostname != "" {
		isNickname, nicknameExists := peer.Nicknames[desc.Hostname]
		if !nicknameExists || !isNickname {
			for _, domain := range e.State.Prefs.LocalDomains {
				if domain != "" && strings.HasSuffix(desc.Hostname, domain) {
					if peer.Nicknames == nil {
						peer.Nicknames = make(map[string]bool)
					}
					peer.Nicknames[desc.Hostname] = true
					break
				}
			}
		}
	}

	elems := make(map[netip.Addr]bool)
	for _, addr := range desc.Addrs {
		if !addr.IsValid() {
			continue
		}
		for _, prefix := range subnets {
			if prefix.IsValid() && prefix.Contains(addr) {
				elems[addr] = true
				break
			}
		}
	}
	for _, addr := range e.State.SystemState.Gateways {
		if _, found := elems[addr]; found {
			e.State.Peers[peerID].UseAsGateway = true
			break
		}
	}

	e.AddTrigger(&syncPeerTrigger{peerID})
	return nil
}

/* Unused functions ported from python version
func (e *organizeTransaction) reject(d *core.Descriptor, reason string) {
	s := schema.ExportDescriptor(d)
	e.AddAction("REJECT", schema.SerializeDescriptorString(&s), reason)
}

func (e *organizeTransaction) log(message string, args ...string) {
	e.AddAction("LOG", args...)
}
*/

func (e *organizeTransaction) ignore(args ...string) {
	e.AddAction("IGNORE", args...)
}
