package organize

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

type organizeTransaction struct {
	eventName string
	eventArgs []string
	State     *core.OrganizeState
	Triggers  []organizeTrigger
	Actions   []organizeActionLog
	Messages  []string
}

func (e *organizeTransaction) AddTrigger(t organizeTrigger) {
	e.Triggers = append(e.Triggers, t)
}

type organizeActionLog struct {
	Name string
	Args []string
}

func (e *organizeTransaction) AddAction(name string, args ...string) {
	e.Actions = append(e.Actions, organizeActionLog{name, args})
}

func (e *organizeTransaction) AddMessage(m string) {
	e.Messages = append(e.Messages, m)
}

type organizeEvent interface {
	eventText() (name string, args []string)
	processEvent(t *organizeTransaction) error
}

type verifyAndPinPeerEvent struct {
	VK       string
	HostName string
}

func (e *verifyAndPinPeerEvent) eventText() (string, []string) {
	return "VERIFY_AND_PIN_PEER", []string{e.VK, e.HostName}
}

func (e *verifyAndPinPeerEvent) processEvent(t *organizeTransaction) error {
	return t.verifyAndPinPeer(e.VK, e.HostName)
}

type userRemovePeerEvent struct {
	Query string
}

func (e *userRemovePeerEvent) eventText() (string, []string) {
	return "USER_REMOVE_PEER", []string{e.Query}
}

func (e *userRemovePeerEvent) processEvent(t *organizeTransaction) error {
	p := peersQuery(t.State.Peers, e.Query)
	if p == nil {
		t.ignore("no such peer")
		return nil
	}
	return t.removePeer(peerID(p))
}

type userPeerAddrAddEvent struct {
	VK    string
	Value string
}

func (e *userPeerAddrAddEvent) eventText() (string, []string) {
	return "USER_PEER_ADDR_ADD", []string{e.VK, e.Value}
}

func (e *userPeerAddrAddEvent) processEvent(t *organizeTransaction) error {
	addr, err := netip.ParseAddr(e.Value)
	if err != nil {
		return err
	}
	return t.peerAddrAdd(e.VK, addr)
}

type userPeerAddrDelEvent struct {
	VK    string
	Value string
}

func (e *userPeerAddrDelEvent) eventText() (string, []string) {
	return "USER_PEER_ADDR_DEL", []string{e.VK, e.Value}
}

func (e *userPeerAddrDelEvent) processEvent(t *organizeTransaction) error {
	addr, err := netip.ParseAddr(e.Value)
	if err != nil {
		return err
	}
	return t.peerAddrDel(e.VK, addr)
}

type editOperation int

const (
	editAdd editOperation = iota + 1
	editRemove
	editSet
)

func (o editOperation) String() string {
	switch o {
	case editAdd:
		return "ADD"
	case editRemove:
		return "REMOVE"
	case editSet:
		return "SET"
	default:
		return ""
	}
}

type userEditEvent struct {
	Operation editOperation
	Path      []string
	Value     any
}

func (e *userEditEvent) eventText() (string, []string) {
	return "USER_EDIT", []string{e.Operation.String(), strings.Join(e.Path, ", "), fmt.Sprintf("%s", e.Value)}
}

func (e *userEditEvent) processEvent(t *organizeTransaction) error {
	return t.edit(e.Operation, e.Path, e.Value)
}

type releaseGatewayEvent struct{}

func (releaseGatewayEvent) eventText() (string, []string) {
	return "RELEASE_GATEWAY", []string{}
}

func (releaseGatewayEvent) processEvent(t *organizeTransaction) error {
	return t.releaseGateway()
}

type newSystemStateEvent struct {
	SystemState *core.SystemState
}

func (e *newSystemStateEvent) eventText() (string, []string) {
	return "NEW_SYSTEM_STATE", []string{fmt.Sprintf("%v", e.SystemState)}
}

func (e *newSystemStateEvent) processEvent(t *organizeTransaction) error {
	return t.adjustToNewSystemState(e.SystemState)
}

type incomingDescriptorEvent struct {
	DescriptorString string
}

func (e *incomingDescriptorEvent) eventText() (string, []string) {
	return "INCOMING_DESCRIPTOR", []string{e.DescriptorString}
}

func (e *incomingDescriptorEvent) processEvent(t *organizeTransaction) error {
	s, err := schema.ParseDescriptorString(e.DescriptorString)
	if err != nil {
		return err
	}
	if !verifyDescriptorSignature(&s) {
		// TODO: should be discarded with no answer?
		core.LogDebug("incoming descriptor was not valid")
		return nil
	}

	d, err := schema.ImportDescriptor(&s)
	if err != nil {
		return err
	}
	return t.incomingDescriptor(&d)
}

type customEvent struct {
	name string
	args []string
	f    func(t *organizeTransaction) error
}

func (e *customEvent) eventText() (string, []string) {
	return e.name, e.args
}

func (e *customEvent) processEvent(t *organizeTransaction) error {
	return e.f(t)
}
