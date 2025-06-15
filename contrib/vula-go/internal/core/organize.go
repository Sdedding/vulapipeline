package core

import (
	"net/netip"
	"time"
)

type OrganizeSync interface {
	Sync(dryRun bool) ([]string, error)
}

type OrganizeDebug interface {
	DumpState(interactive bool) (string, error)
	TestAuth(interactive bool) (string, error)
}

type OrganizePeers interface {
	ShowPeer(query string) (string, error)
	PeerDescriptor(query string) (string, error)
	PeerIds(which string) ([]string, error)
	Rediscover() (string, error)
	SetPeer(vk string, path []string, value string) (string, error)
	RemovePeer(vk string) (string, error)
	PeerAddrAdd(vk string, value string) (string, error)
	PeerAddrDel(vk string, value string) (string, error)
	OurLatestDescriptors() (string, error)
	GetVkByName(hostname string) (string, error)
	VerifyAndPinPeer(vk string, hostname string) (string, error)
}

type OrganizeProcessDescriptor interface {
	ProcessDescriptorString(descriptor string) (string, error)
}

type OrganizePrefs interface {
	ShowPrefs() (string, error)
	SetPref(pref string, value string) (string, error)
	AddPref(pref string, value string) (string, error)
	RemovePref(pref string, value string) (string, error)
	ReleaseGateway() (string, error)
}

type Organize interface {
	OrganizeSync
	OrganizeDebug
	OrganizePeers
	OrganizeProcessDescriptor
	OrganizePrefs
}

type OrganizeState struct {
	Prefs       *Prefs
	Peers       map[string]*Peer
	SystemState *SystemState
	EventLog    []string
}

type OrganizeStateRepository interface {
	LoadOrganizeState() (*OrganizeState, error)
	SaveOrganizeState(state *OrganizeState) error
}

type Prefs struct {
	PinNewPeers        bool
	AutoRepair         bool
	SubnetsAllowed     []netip.Prefix
	SubnetsForbidden   []netip.Prefix
	IfacePrefixAllowed []string
	AcceptNonlocal     bool
	LocalDomains       []string
	EphemeralMode      bool
	AcceptDefaultRoute bool
	OverwriteUnpinned  bool
	ExpireTime         time.Duration
	PrimaryIP          netip.Addr
	RecordEvents       bool
	EnableIPv6         bool
	EnableIPv4         bool
}

// SystemState object stores the parts of the system's state which are
// relevant to events in the organize state engine. The object should be
// updated (replaced) whenever these values change, via the
// event_NEW_SYSTEM_STATE event.
type SystemState struct {
	CurrentSubnets    map[netip.Prefix][]netip.Addr
	CurrentInterfaces map[string][]netip.Addr
	OurWgPK           []byte
	Gateways          []netip.Addr
	HasV6             bool
}

type HostsFileRepository interface {
	WriteHostsFile(entries [][2]string) error
}
