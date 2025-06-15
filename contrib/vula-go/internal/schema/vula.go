package schema

import (
	"fmt"
	"slices"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

type Descriptor struct {
	P        IPAddr       `yaml:"p,omitempty"`
	V4A      IPAddrList   `yaml:"v4a,omitempty"`
	V6A      IPAddrList   `yaml:"v6a,omitempty"`
	PK       Base64       `yaml:"pk"`
	C        Base64       `yaml:"c"`
	Hostname string       `yaml:"hostname"`
	Port     int64        `yaml:"port"`
	VK       Base64       `yaml:"vk"`
	DT       int64        `yaml:"dt"`
	VF       int64        `yaml:"vf"`
	R        IPPrefixList `yaml:"r"`
	E        bool         `yaml:"e"`
	S        Base64       `yaml:"s,omitempty"`
}

type Peer struct {
	Descriptor   Descriptor        `yaml:"descriptor"`
	Petname      string            `yaml:"petname"`
	Nicknames    Map[string, bool] `yaml:"nicknames"`
	IPv4Addrs    Map[IPAddr, bool] `yaml:"IPv4addrs"`
	IPv6Addrs    Map[IPAddr, bool] `yaml:"IPv6addrs"`
	Enabled      bool              `yaml:"enabled"`
	Verified     bool              `yaml:"verified"`
	Pinned       bool              `yaml:"pinned"`
	UseAsGateway bool              `yaml:"use_as_gateway"`
}

type SystemState struct {
	CurrentSubnets    Map[IPPrefix, Slice[IPAddr]] `yaml:"current_subnets"`
	CurrentInterfaces Map[string, Slice[IPAddr]]   `yaml:"current_interfaces"`
	OurWgPK           Base64                       `yaml:"our_wg_pk"`
	Gateways          Slice[IPAddr]                `yaml:"gateways"`
	HasV6             bool                         `yaml:"has_v6"`
}

type OrganizeState struct {
	Prefs       Prefs             `yaml:"prefs"`
	Peers       Map[string, Peer] `yaml:"peers"`
	SystemState SystemState       `yaml:"system_state"`
	EventLog    []string          `yaml:"event_log"`
}

type Prefs struct {
	PinNewPeers        bool            `yaml:"pin_new_peers"`
	AutoRepair         bool            `yaml:"auto_repair"`
	SubnetsAllowed     Slice[IPPrefix] `yaml:"subnets_allowed"`
	SubnetsForbidden   Slice[IPPrefix] `yaml:"subnets_forbidden"`
	IfacePrefixAllowed Slice[string]   `yaml:"iface_prefix_allowed"`
	AcceptNonlocal     bool            `yaml:"accept_nonlocal"`
	LocalDomains       Slice[string]   `yaml:"local_domains"`
	EphemeralMode      bool            `yaml:"ephemeral_mode"`
	AcceptDefaultRoute bool            `yaml:"accept_default_route"`
	OverwriteUnpinned  bool            `yaml:"overwrite_unpinned"`
	ExpireTime         int64           `yaml:"expire_time"`
	PrimaryIP          IPAddr          `yaml:"primary_ip"`
	RecordEvents       bool            `yaml:"record_events"`
	EnableIPv6         bool            `yaml:"enable_ipv6"`
	EnableIPv4         bool            `yaml:"enable_ipv4"`
}

type KeyFile struct {
	PqCtidhP512SecKey  Base64 `yaml:"pq_ctidhP512_sec_key"`
	PqCtidhP512PubKey  Base64 `yaml:"pq_ctidhP512_pub_key"`
	VkEd25519SecKey    Base64 `yaml:"vk_Ed25519_sec_key"`
	VkEd25519PubKey    Base64 `yaml:"vk_Ed25519_pub_key"`
	WgCurve25519SecKey Base64 `yaml:"wg_Curve25519_sec_key"`
	WgCurve25519PubKey Base64 `yaml:"wg_Curve25519_pub_key"`
}

type Result struct {
	Event   string        `yaml:"event"`
	Actions Slice[string] `yaml:"actions"`
	Writes  []string      `yaml:"writes"`
}

func ParseDescriptorString(s string) (d Descriptor, err error) {
	fields := strings.Split(s, ";")
	items := []StringMapItem{}

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		sepIndex := strings.IndexByte(field, '=')
		if sepIndex < 0 {
			err = &core.ErrParse{
				Err: fmt.Errorf("field without seperator: %s", field),
			}
			return
		}

		key := strings.TrimSpace(field[:sepIndex])
		value := strings.TrimSpace(field[sepIndex+1:])

		items = append(items, StringMapItem{Key: key, Value: value})
	}
	err = Write(items, nil, &d)
	return
}

func SerializeDescriptorString(d *Descriptor) string {
	s := []byte{}
	items := []StringMapItem{}
	err := Read(d, nil, &items)
	if err != nil {
		panic("descriptor serialization should not fail")
	}

	slices.SortFunc(items, func(a, b StringMapItem) int {
		return strings.Compare(a.Key, b.Key)
	})

	for _, item := range items {
		s = append(s, item.Key...)
		s = append(s, '=')
		s = append(s, item.Value...)
		s = append(s, ';')
		s = append(s, ' ')
	}
	if len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}

	return string(s)
}
