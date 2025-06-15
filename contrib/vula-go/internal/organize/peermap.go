package organize

import (
	"bytes"
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"errors"
	"fmt"
	"net/netip"
	"slices"
)

var (
	errNotFound = errors.New("vula: item not found")
)

func peersfindUnique(ps map[string]*core.Peer, pred func(p *core.Peer) bool) (peer *core.Peer, err error) {
	for _, p := range ps {
		if pred(p) {
			// we assume that peers are unique
			peer = p
			return
		}
	}
	err = errNotFound
	return
}

func peersWithHostname(ps map[string]*core.Peer, name string) (*core.Peer, error) {
	peer, err := peersfindUnique(ps, func(p *core.Peer) bool {
		_, ok := p.Nicknames[name]
		return ok && p.Enabled
	})

	switch err {
	case errNotFound:
		err = fmt.Errorf("%w: peer with hostname '%s'", err, name)
	}

	return peer, err
}

func peersGateways(ps map[string]*core.Peer) []*core.Peer {
	gateways := []*core.Peer{}
	for _, peer := range ps {
		if peer.UseAsGateway {
			gateways = append(gateways, peer)
		}
	}
	return gateways
}

func peersGetWithEnabledIP(ps map[string]*core.Peer, ip netip.Addr) (*core.Peer, bool) {
	for _, peer := range ps {
		if peer.Addrs[ip] {
			return peer, true
		}
	}
	return nil, false
}

func peersFilter(ps map[string]*core.Peer, filter func(*core.Peer) bool) map[string]*core.Peer {
	filtered := map[string]*core.Peer{}
	for k, peer := range ps {
		if filter(peer) {
			filtered[k] = peer
		}
	}
	return filtered
}

func peersGetEnabled(ps map[string]*core.Peer) map[string]*core.Peer {
	return peersFilter(ps, func(p *core.Peer) bool {
		return p.Enabled
	})
}

func peersGetDisabled(ps map[string]*core.Peer) map[string]*core.Peer {
	return peersFilter(ps, func(p *core.Peer) bool {
		return !p.Enabled
	})
}

func peersIDs(ps map[string]*core.Peer) []string {
	var ids []string
	for _, peer := range ps {
		ids = append(ids, peerID(peer))
	}
	return ids
}

/*
// Conflicts returns comma-separated list of colliding peer ids
func peersConflicts(ps map[string]*core.Peer) string {
	enabledGateways := []*core.Peer{}
	for _, p := range ps {
		if p.Enabled && p.UseAsGateway {
			enabledGateways = append(enabledGateways, p)
		}
	}

	peerIDs := []string{}
	for _, p := range ps {
		if !p.Enabled {
			continue
		}

		conflicts := peersConflictsForDescriptor(ps, &p.Descriptor)
		if len(conflicts) > 0 {
			peerIDs = append(peerIDs, peerID(p))
		}
	}

	if len(enabledGateways) > 1 {
		for _, p := range enabledGateways {
			peerIDs = append(peerIDs, peerID(p))
		}
	}

	return strings.Join(peerIDs, ",")
}
*/

// ConflictsForDescriptor returns list of enabled peers a descriptor
// has a conflicting wg_pk, hostname, or IP address with (ignoring itself).
func peersConflictsForDescriptor(ps map[string]*core.Peer, d *core.Descriptor) []*core.Peer {
	peers := map[string]*core.Peer{}
	for _, p := range ps {
		if !p.Enabled || bytes.Equal(p.Descriptor.VulaPK, d.VulaPK) {
			continue
		}

		if slices.Contains(peerEnabledNames(p), d.Hostname) {
			peers[peerID(p)] = p
		}

		if slices.Equal(p.Descriptor.WireGuardPK, d.WireGuardPK) {
			peers[peerID(p)] = p
		}

		ipAddrs := map[netip.Addr]bool{}
		for ip := range p.Addrs {
			ipAddrs[ip] = true
		}

		for _, ip := range d.Addrs {
			if ipAddrs[ip] {
				peers[peerID(p)] = p
				break
			}
		}
	}

	s := make([]*core.Peer, 0, len(peers))
	for _, p := range peers {
		s = append(s, p)
	}
	return s
}

// Query returns peer by vk, hostname, or IP. Nil if no match.
func peersQuery(ps map[string]*core.Peer, query string) *core.Peer {
	peers := []*core.Peer{}
	found := false
	for _, p := range ps {
		if peerID(p) == query {
			peers = append(peers, p)
			found = true
		}
	}

	if !found {
		for _, p := range ps {
			if p.Enabled && slices.Contains(peerEnabledNames(p), query) {
				peers = append(peers, p)
				found = true
			}
		}
	}

	if !found {
		for _, p := range ps {
			if p.Enabled && slices.Contains(peerEnabledIPsString(p), query) {
				peers = append(peers, p)
				found = true
			}
		}
	}

	if len(peers) > 1 {
		// this should not be possible
		panic("vula: query for peers should not yield more than one item")
	}

	if len(peers) == 1 {
		return peers[0]
	}
	return nil
}

func hostsFileEntries(ps map[string]*core.Peer) [][2]string {
	s := [][2]string{}
	for _, peer := range peersGetEnabled(ps) {
		for _, name := range peerEnabledNames(peer) {
			if len(peer.Descriptor.Addrs) > 0 {
				s = append(s, [2]string{peer.Descriptor.Addrs[0].String(), name})
			}
		}
	}
	return s
}
