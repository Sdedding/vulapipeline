package organize

import (
	"net/netip"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

type organizeTrigger interface {
	triggerName() string
	processTrigger(o *O) error
}

type syncPeerTrigger struct {
	ID string
}

func (t *syncPeerTrigger) triggerName() string {
	return "SyncPeer"
}

func (t *syncPeerTrigger) processTrigger(o *O) error {
	_, err := o.syncPeer(o.state.Peers, t.ID, o.state.SystemState, o.config.RoutingTable, false)
	return err
}

type removeRoutesTrigger struct {
	Routes []netip.Prefix
}

func (t *removeRoutesTrigger) triggerName() string {
	return "RemoveRoutes"
}

func (t *removeRoutesTrigger) processTrigger(o *O) error {
	_, err := o.config.NetworkSystem.RemoveRoutes(t.Routes, o.config.RoutingTable, o.wgi.Name(), false)
	return err
}

type getNewSystemStateTrigger struct{}

func (t *getNewSystemStateTrigger) triggerName() string {
	return "GetNewSystemState"
}

func (t *getNewSystemStateTrigger) processTrigger(o *O) error {
	go func() {
		_, err := o.getNewSystemState("")
		if err != nil {
			core.LogWarn(err)
		}
	}()
	return nil
}

type removeWgPeerTrigger struct {
	WgPK []byte
}

func (t *removeWgPeerTrigger) triggerName() string {
	return "RemoveWgPeer"
}

func (t *removeWgPeerTrigger) processTrigger(o *O) error {
	log, err := o.wgi.SetPeer(&core.PeerConfig{PublicKey: t.WgPK, Remove: true}, false)
	core.LogDebug("REMOVE PEER LOG ", log)
	return err
}

type removeUnknownTrigger struct{}

func (t *removeUnknownTrigger) triggerName() string {
	return "RemoveUnknown"
}

func (t *removeUnknownTrigger) processTrigger(o *O) error {
	_, err := o.removeUnknown(o.config.RoutingTable, o.state.Peers, false)
	return err
}
