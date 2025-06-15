package organize

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"codeberg.org/vula/highctidh/src/ctidh512"
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"gopkg.in/yaml.v3"
)

// systemStateCurrentIPs returns the currently assigned IP addresses of the system
func systemStateCurrentIPs(s *core.SystemState) []netip.Addr {
	var ips []netip.Addr

	for _, subnet := range s.CurrentSubnets {
		ips = append(ips, subnet...)
	}

	return ips
}

func systemStateCurrentSubnetsNoULA(s *core.SystemState) map[netip.Prefix][]netip.Addr {
	subnets := make(map[netip.Prefix][]netip.Addr, len(s.CurrentSubnets))

	if s.CurrentSubnets == nil {
		return subnets
	}

	for prefix, subnet := range s.CurrentSubnets {
		if prefix != core.VulaSubnet {
			subnets[prefix] = subnet
		}
	}

	return subnets
}

func (e *organizeTransaction) verifyAndPinPeer(vk, hostname string) error {
	e.AddAction("VerifyAndPinPeer", vk, hostname)

	peer, err := peersWithHostname(e.State.Peers, hostname)
	if err != nil {
		return err
	}

	id := peerID(peer)
	if vk != id {
		return fmt.Errorf("expecte %s: expected: %s have: %s", hostname, vk, id)
	}

	peer.Verified = true
	peer.Pinned = true
	return nil
}

func (e *organizeTransaction) releaseGateway() error {
	e.AddAction("ReleaseGateway")

	gateways := peersGateways(e.State.Peers)
	if len(gateways) == 0 {
		e.AddMessage("no current gateway peer")
		return nil
	}

	gateways[0].UseAsGateway = false
	e.AddTrigger(&getNewSystemStateTrigger{})
	return nil
}

func getOverlap[T comparable](a, b []T) []T {
	ma := make(map[T]struct{}, len(a))
	for _, v := range a {
		ma[v] = struct{}{}
	}

	s := []T{}
	for _, v := range b {
		_, ok := ma[v]
		if ok {
			s = append(s, v)
		}
	}

	return s
}

func (e *organizeTransaction) incomingDescriptor(d *core.Descriptor) error {
	e.AddAction("IncomingDescriptor")

	if bytes.Equal(d.WireGuardPK, e.State.SystemState.OurWgPK) {
		e.AddMessage(fmt.Sprintf("ignore descriptor: %s: has our wg pk", descriptorToString(d)))
		return nil
	}

	if !checkDescriptorFreshness(d, time.Now()) {
		e.AddMessage(fmt.Sprintf("reject descriptor: %s: timestamp too old", descriptorToString(d)))
		return nil
	}

	existingPeer := e.State.Peers[descriptorID(d)]
	if existingPeer != nil && d.ValidStart <= existingPeer.Descriptor.ValidStart {
		e.AddMessage(fmt.Sprintf("ignore descriptor: %s: replay", descriptorToString(d)))
		return nil
	}

	if len(addrsInSubnets(d.Addrs, e.State.SystemState.CurrentSubnets)) == 0 {
		e.AddMessage(fmt.Sprintf("reject descriptor: %s: wrong subnet, current subnets are %v", descriptorToString(d), e.State.SystemState.CurrentSubnets))
		return nil
	}

	hostnameEndsWithDomain := false
	for _, domain := range e.State.Prefs.LocalDomains {
		if strings.HasSuffix(d.Hostname, domain) {
			hostnameEndsWithDomain = true
			break
		}
	}
	if !hostnameEndsWithDomain {
		e.AddMessage(fmt.Sprintf("reject descriptor: %s: invalid domain", descriptorToString(d)))
		return nil
	}

	conflictingPeers := peersConflictsForDescriptor(e.State.Peers, d)
	if len(conflictingPeers) > 0 {
		for _, peer := range conflictingPeers {
			if peer.Pinned {
				e.AddMessage(fmt.Sprintf("reject descriptor: %s: conflict %v", descriptorToString(d), peer))
				return nil
			}
		}

		for _, peer := range conflictingPeers {
			err := e.removePeer(peerID(peer))
			if err != nil {
				return err
			}
		}
	}

	if existingPeer != nil {
		return e.updatePeerDescriptor(peerID(existingPeer), d)
	}
	return e.acceptNewPeer(d)
}

type Config struct {
	// dependencies
	StateRepository     core.OrganizeStateRepository
	HostsFileRepository core.HostsFileRepository
	KeyRepository       core.KeyRepository
	Discover            core.Discover
	Publish             core.Publish
	NetworkSystem       core.NetworkSystem
	WireguardSystem     core.WireguardSystem

	// configuration values
	InterfaceName  string
	Port           uint16
	FirewallMark   int
	RoutingTable   int
	IPRulePriority int
}

type O struct {
	mux    sync.Mutex
	state  *core.OrganizeState
	keys   *core.Keys
	wgi    core.WireguardInterface
	dh     *ctidh512Impl
	config *Config

	events chan queueEvent
	done   chan struct{}
}

func New(config *Config) (o *O, err error) {
	// load keys
	keys, err := generateOrReadKeys(config.KeyRepository)
	if err != nil {
		core.LogWarn("failed to load keys:", err)
		keys, err = genKeys()
		if err != nil {
			return
		}
	}

	ctidhSk := ctidh512.NewEmptyPrivateKey()
	err = ctidhSk.FromBytes(keys.PqCtidhP512.SK)
	if err != nil {
		return nil, err
	}

	// create wiregard interface
	wgi := config.WireguardSystem.GetInterface(config.InterfaceName)

	dh := newCtidh512Impl(ctidhSk)

	state, err := loadState(config.StateRepository)
	if err != nil {
		return nil, fmt.Errorf("loading organize state: %w", err)
	}

	if state.SystemState == nil {
		state.SystemState = defaultSystemState
	}
	state.SystemState.OurWgPK = keys.WgCurve25519.PK

	o = &O{
		state:  state,
		keys:   keys,
		dh:     dh,
		wgi:    wgi,
		config: config,

		events: make(chan queueEvent, 256),
		done:   make(chan struct{}),
	}

	// start the event loop
	go o.processEvents()

	err = o.initialize()
	if err != nil {
		err := o.Close()
		if err != nil {
			return nil, err
		}
	}
	return
}

func (o *O) Close() (err error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if o.events != nil {
		close(o.events)
		o.events = nil
	}
	if o.done != nil {
		close(o.done)
	}
	return
}

func (o *O) save() error {
	if err := o.config.StateRepository.SaveOrganizeState(o.state); err != nil {
		return err
	}

	if err := o.config.HostsFileRepository.WriteHostsFile(hostsFileEntries(o.state.Peers)); err != nil {
		return err
	}
	return nil
}

type queueEvent struct {
	event organizeEvent
	done  chan<- eventResult
}

type eventResult struct {
	err     error
	state   *core.OrganizeState
	details eventResultDetails
}

type eventResultDetails struct {
	event   string
	args    []string
	actions []organizeActionLog
}

func (r *eventResult) serialize() (string, error) {
	actions := make([]string, len(r.details.actions))
	for i, a := range r.details.actions {
		actions[i] = fmt.Sprintf("%s [%s]", a.Name, strings.Join(a.Args, ", "))
	}

	s := schema.Result{
		Event:   fmt.Sprintf("%s [%s]", r.details.event, strings.Join(r.details.args, ", ")),
		Actions: actions,
		Writes:  []string{},
	}
	data, _ := yaml.Marshal(s)
	return string(data), r.err
}

func (o *O) processEvent(t *organizeTransaction, event organizeEvent) (err error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	err = event.processEvent(t)
	if err == nil {
		// update state
		o.state = t.State
		err = o.save()
	}
	return
}

func (o *O) processEventTriggers(triggers []organizeTrigger) (err error) {
	compoundErrorBuilder := compoundErrorBuilder{}
	for _, trigger := range triggers {
		err = trigger.processTrigger(o)
		compoundErrorBuilder.Add(err)
	}

	return compoundErrorBuilder.Build()
}

func (o *O) processEvents() {
	for event := range o.events {
		name, args := event.event.eventText()
		t := &organizeTransaction{
			eventName: name,
			eventArgs: args,
			State:     o.state.DeepCopy(),
		}

		err := o.processEvent(t, event.event)

		for _, m := range t.Messages {
			core.LogDebug(m)
		}

		details := eventResultDetails{t.eventName, t.eventArgs, t.Actions}
		event.done <- eventResult{err, o.state, details}
		if err != nil {
			core.LogWarn(err)
			continue
		}

		err = o.processEventTriggers(t.Triggers)
		if err != nil {
			core.LogWarn(err)
		}
	}
}

func (o *O) queueEvent(event organizeEvent) <-chan eventResult {
	ch := make(chan eventResult)
	o.events <- queueEvent{event, ch}
	return ch
}

func (o *O) Sync(dryRun bool) (log []string, err error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	syncLog, err := o.syncInterfaces(o.keys.WgCurve25519.SK, o.config.Port, o.config.FirewallMark, o.state.Prefs.PrimaryIP, dryRun)
	log = append(log, syncLog...)
	if err != nil {
		err = fmt.Errorf("failed to sync wg interface: %w", err)
		return
	}

	syncLog, err = o.config.NetworkSystem.SyncRules(o.config.RoutingTable, o.config.FirewallMark, o.config.IPRulePriority, dryRun)
	log = append(log, syncLog...)
	if err != nil {
		err = fmt.Errorf("failed to sync ip rules: %w", err)
		return
	}

	for id, p := range o.state.Peers {
		core.LogDebugf("syncing peer %s", peerNameAndID(p))

		syncLog, err = o.syncPeer(o.state.Peers, id, o.state.SystemState, o.config.RoutingTable, dryRun)
		if err != nil {
			syncLog = append(syncLog, fmt.Sprintf("failed to sync peer: %s", err.Error()))
		}
		log = append(log, syncLog...)
	}

	removeLog, err := o.removeUnknown(o.config.RoutingTable, o.state.Peers, dryRun)
	log = append(log, removeLog...)
	return
}

func (o *O) DumpState(interactive bool) (string, error) {
	// TODO: implement
	return "", nil
}

func (o *O) TestAuth(interactive bool) (string, error) {
	// TODO: implement
	return "", nil
}

func (o *O) ShowPeer(query string) (string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	peer := peersQuery(o.state.Peers, query)
	if peer == nil {
		return "", fmt.Errorf("no peer matched query %s", query)
	}
	return string(ShowPeer(peer, &core.PeerStats{
		// TODO: Get actual stats from sys
		LatestHandshake:    0,
		RxBytes:            0,
		TxBytes:            0,
		HasLatestHandshake: false,
	}).ToConsole()), nil
}

func (o *O) PeerDescriptor(query string) (string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	peer := peersQuery(o.state.Peers, query)
	if peer == nil {
		return "", fmt.Errorf("no peer with query %s", query)
	}
	// TODO: perhaps here the format is not exactly the same as in the python `str(peer.descriptor)`
	s := schema.ExportDescriptor(&peer.Descriptor)
	return schema.SerializeDescriptorString(&s), nil
}

func (o *O) PeerIds(which string) ([]string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	var peers map[string]*core.Peer

	if which == "enabled" {
		peers = peersGetEnabled(o.state.Peers)
	} else if which == "disabled" {
		peers = peersGetDisabled(o.state.Peers)
	} else if which == "all" {
		peers = o.state.Peers
	} else {
		return nil, fmt.Errorf("unknown filter \"%s\"", which)
	}

	return peersIDs(peers), nil
}

func (o *O) Rediscover() (string, error) {
	// TODO: implement
	return "", nil
}

func (o *O) SetPeer(vk string, path []string, value string) (string, error) {
	result := <-o.queueEvent(&userEditEvent{editSet, append([]string{"peers", vk}, path...), value})
	return result.serialize()
}

func (o *O) RemovePeer(vk string) (string, error) {
	result := <-o.queueEvent(&userRemovePeerEvent{vk})
	return result.serialize()
}

func (o *O) PeerAddrAdd(vk string, value string) (string, error) {
	result := <-o.queueEvent(&userPeerAddrAddEvent{vk, value})
	return result.serialize()
}

func (o *O) PeerAddrDel(vk string, value string) (string, error) {
	result := <-o.queueEvent(&userPeerAddrDelEvent{vk, value})
	return result.serialize()
}

func (o *O) OurLatestDescriptors() (string, error) {
	// TODO: implement, returns latest descriptors as json
	return "", nil
}

func (o *O) GetVkByName(hostname string) (string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	peer, err := peersWithHostname(o.state.Peers, hostname)
	if err != nil {
		return "", err
	}
	return peerID(peer), nil
}

func (o *O) VerifyAndPinPeer(vk string, hostname string) (string, error) {
	result := <-o.queueEvent(&verifyAndPinPeerEvent{vk, hostname})
	return result.serialize()
}

func (o *O) ProcessDescriptorString(descriptor string) (string, error) {
	core.LogDebugf("discovered descriptor: %s", descriptor)
	result := <-o.queueEvent(&incomingDescriptorEvent{descriptor})
	return result.serialize()
}

func (o *O) ShowPrefs() (string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	prefs := schema.ExportPrefs(o.state.Prefs)
	serializedPrefs, err := yaml.Marshal(prefs)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", string(serializedPrefs)), nil
}

func (o *O) SetPref(pref string, value string) (string, error) {
	result := <-o.queueEvent(&userEditEvent{editSet, []string{"prefs", pref}, value})
	return result.serialize()
	/*
		o.mux.Lock()
		defer o.mux.Unlock()

		prefs := exportPrefs(o.state.Prefs)
		err := schema.Write(value, []string{pref}, &prefs)
		// verify
		if err == nil {
			updated := importPrefs(&prefs)
			o.state.Prefs = &updated
		}
		// TODO: return something that makes sense. For the python version of vula it would be a `Result`.
		return "", err
	*/
}

func (o *O) AddPref(pref string, value string) (string, error) {
	result := <-o.queueEvent(&userEditEvent{editAdd, []string{"prefs", pref}, value})
	// TODO: return something that makes sense. For the python version of vula it would be a `Result`.
	return result.serialize()
}

func (o *O) RemovePref(pref string, value string) (string, error) {
	result := <-o.queueEvent(&userEditEvent{editRemove, []string{"prefs", pref}, value})
	// TODO: return something that makes sense. For the python version of vula it would be a `Result`.
	return result.serialize()
}

func (o *O) ReleaseGateway() (string, error) {
	result := <-o.queueEvent(&releaseGatewayEvent{})
	// TODO: return serialized representation of state (yaml)
	return result.serialize()
}

func (o *O) getSystemStateConfig() systemStateConfig {
	o.mux.Lock()
	defer o.mux.Unlock()

	config := systemStateConfig{
		WireguardPK:        o.state.SystemState.OurWgPK,
		IfacePrefixAllowed: o.state.Prefs.IfacePrefixAllowed,
		PrimaryIP:          o.state.Prefs.PrimaryIP,
		SubnetsForbidden:   o.state.Prefs.SubnetsForbidden,
		EnableIPv4:         o.state.Prefs.EnableIPv4,
		EnableIPv6:         o.state.Prefs.EnableIPv6,
	}
	return config
}

func (o *O) getNewSystemState(reason string) (log []string, err error) {

	config := o.getSystemStateConfig()
	networkSystemState, err := o.config.NetworkSystem.GetSystemState(config.EnableIPv4, config.EnableIPv6, config.IfacePrefixAllowed, config.SubnetsForbidden, config.PrimaryIP)
	if err != nil {
		return
	}

	_, wireguardPK, _, _, err := o.wgi.Configuration()

	newState := &core.SystemState{
		CurrentSubnets:    networkSystemState.CurrentSubnets,
		CurrentInterfaces: networkSystemState.CurrentInterfaces,
		OurWgPK:           wireguardPK,
		Gateways:          networkSystemState.Gateways,
		HasV6:             networkSystemState.HasV6,
	}

	//TODO: check if state was changed
	core.LogInfof("Checked system state because %s; found changes, running sync/repair", reason)
	result := <-o.queueEvent(&newSystemStateEvent{newState})
	if result.err != nil {
		err = fmt.Errorf("fatal unable to handle new system state: %w", err)
		return
	}

	// FIXME: ensure that triggers do everything necessary, and then
	// remove this full repair sync call here:
	_, err = o.Sync(false)
	return

}

func (o *O) processNetlinkUpdates(updates <-chan string) {
	for update := range updates {
		_, err := o.getNewSystemState(fmt.Sprintf("netlink event: %s", update))
		if err != nil {
			core.LogWarn(err)
		}
	}
}

var (
	ipv4LL  = netip.MustParsePrefix("169.254.0.0/16")
	ipv6LL  = netip.MustParsePrefix("fe80::/10")
	ipv6ULA = netip.MustParsePrefix("fc00::/7")
)

// initialize initializes the O instance
// it is required that the event loop is already running
func (o *O) initialize() error {
	err := o.config.Discover.Listen(nil, "")
	if err != nil {
		return err
	}

	state := func() *core.OrganizeState {
		o.mux.Lock()
		defer o.mux.Unlock()
		return o.state
	}()

	if !state.Prefs.PrimaryIP.IsValid() {
		s := core.VulaSubnet.Addr().AsSlice()
		_, err := io.ReadFull(rand.Reader, s[6:])
		if err != nil {
			return err
		}
		addr, ok := netip.AddrFromSlice(s)
		if !ok {
			return fmt.Errorf("failed to create random address")
		}

		event := &customEvent{
			name: "SET_PRIMARY_IP",
			args: []string{addr.String()},
			f: func(t *organizeTransaction) error {
				t.State.Prefs.PrimaryIP = addr
				t.AddTrigger(&getNewSystemStateTrigger{})
				return nil
			},
		}

		result := <-o.queueEvent(event)
		if result.err != nil {
			return result.err
		}

		if !state.Prefs.EnableIPv6 {
			message := "FIXME: v4-only hosts currently not " +
				"supported. preliminary v4-only testing " +
				"involves manually setting the primary_ip " +
				"pref which will avoid hitting this " +
				"exception"
			return fmt.Errorf("%s", message)
		}
	}

	_, err = o.getNewSystemState("")
	if err != nil {
		return err
	}
	o.mux.Lock()
	state = o.state
	o.mux.Unlock()

	if state.Prefs.EnableIPv6 {
		// we always want these two when v6 is enabled. they're in the
		// default prefs, but we add them here to handle upgrading from a
		// pre-v6 prefs file.

		event := &customEvent{
			name: "ADD_IPV6_ALLOWED_SUBNETS",
			args: []string{ipv6LL.String(), ipv6ULA.String()},
			f: func(t *organizeTransaction) error {
				if !slices.Contains(t.State.Prefs.SubnetsAllowed, ipv6LL) {
					t.State.Prefs.SubnetsAllowed = append(t.State.Prefs.SubnetsAllowed, ipv6LL)
				}
				if !slices.Contains(t.State.Prefs.SubnetsAllowed, ipv6ULA) {
					t.State.Prefs.SubnetsAllowed = append(t.State.Prefs.SubnetsAllowed, ipv6ULA)
				}
				return nil
			},
		}

		result := <-o.queueEvent(event)
		if result.err != nil {
			return err
		}
	}

	netlinkUpdates := make(chan string, 1)
	err = o.config.NetworkSystem.StartMonitor(o.done, netlinkUpdates)
	if err != nil {
		return err
	}

	go o.processNetlinkUpdates(netlinkUpdates)

	err = instructZeroconf(o.config, o.keys, state)
	if err != nil {
		return err
	}

	_, err = o.Sync(false)
	return err
}

func instructZeroconf(config *Config, keys *core.Keys, state *core.OrganizeState) error {
	vf := time.Now().UTC().Unix()
	discoverIPs := []netip.Addr{}
	publishIPs := []netip.Addr{}
	descriptors := map[string]string{}

	for iface, addrs := range state.SystemState.CurrentInterfaces {
		allowedAddrs := filterAddrsByPrefix(addrs, state.Prefs.SubnetsAllowed)
		if len(allowedAddrs) == 0 {
			continue
		}

		discoverIPs = append(discoverIPs, allowedAddrs...)
		sortLLFirst(allowedAddrs)
		publishIPs = append(publishIPs, allowedAddrs...)

		// an addr in subnets_forbidden is filtered already when the system state is created
		d := constructServiceDescriptor(keys, state, uint16(config.Port), allowedAddrs, vf)
		s := schema.ExportDescriptor(&d)
		err := signDescriptor(&s, keys.VkEd25519.SK)
		if err != nil {
			return err
		}
		descriptors[iface] = schema.SerializeDescriptorString(&s)
	}

	discoverIPStrings := make([]string, len(discoverIPs))
	for i, addr := range discoverIPs {
		discoverIPStrings[i] = addr.String()
	}

	publishIPStrings := make([]string, len(publishIPs))
	for i, addr := range publishIPs {
		publishIPStrings[i] = addr.String()
	}

	core.LogInfof("discovering on %v and publishing %v", strings.Join(discoverIPStrings, ", "), strings.Join(publishIPStrings, ", "))
	err := config.Discover.Listen(discoverIPStrings, "")
	if err != nil {
		return err
	}
	err = config.Publish.Listen(descriptors)
	if err != nil {
		return err
	}

	core.LogInfof("Current IP(s): %s", strings.Join(publishIPStrings, ", "))
	core.LogDebugf("Current descriptors: %v", descriptors)
	return nil
}

func filterAddrsByPrefix(addrs []netip.Addr, prefixes []netip.Prefix) []netip.Addr {
	filtered := []netip.Addr{}
	for _, addr := range addrs {
		for _, prefix := range prefixes {
			if prefix.Contains(addr) {
				filtered = append(filtered, addr)
				break
			}
		}
	}
	return filtered
}
