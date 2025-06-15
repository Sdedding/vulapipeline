package schema

import (
	"fmt"
	"net/netip"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	IPAddrList   []IPAddr
	IPPrefixList []IPPrefix
)

var (
	_ yaml.Marshaler   = IPAddrList{}
	_ yaml.Unmarshaler = &IPAddrList{}
	_ Reader           = IPAddrList{}
	_ Writer           = &IPAddrList{}
	_ PathReader       = IPAddrList{}
	_ PathWriter       = &IPAddrList{}
	_ PathAdder        = &IPAddrList{}
	_ PathRemover      = &IPAddrList{}
)

func writeCommaSeperated[S ~[]T, T any](v string, t *S) (err error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}

	parts := strings.Split(v, ",")
	s := make([]T, len(parts))
	*t = s
	for i := range parts {
		err = writeValue(parts[i], &s[i])
		if err != nil {
			return
		}
	}
	return
}

func formatCommaSeperated[S ~[]T, T fmt.Stringer](v S) string {
	s := []byte{}
	if len(v) == 0 {
		return ""
	}

	s = append(s, v[0].String()...)
	for i := 1; i < len(v); i++ {
		s = append(s, ',')
		s = append(s, v[i].String()...)
	}
	return string(s)
}

func (v IPAddrList) MarshalYAML() (any, error) {
	return formatCommaSeperated(v), nil
}

func (t *IPAddrList) UnmarshalYAML(n *yaml.Node) error {
	s := ""
	err := n.Decode(&s)
	if err != nil {
		return err
	}
	err = writeCommaSeperated(s, t)
	return err
}

func (v IPAddrList) Read(t any) (err error) {
	switch t := t.(type) {
	case *string:
		*t = formatCommaSeperated(v)
	default:
		err = ReadSlice(v, t)
	}
	return
}

func (t *IPAddrList) Write(v any) (err error) {
	switch v := v.(type) {
	case string:
		err = writeCommaSeperated(v, t)
	default:
		err = WriteSlice(v, t)
	}
	return
}

func (v IPAddrList) ReadPath(path []string, t any) error {
	return ReadSlice(v, t)
}

func (t *IPAddrList) WritePath(v any, path []string) error {
	return WriteSlice(v, t)
}

func (t *IPAddrList) AddPath(v any, path []string) error {
	return SliceAddPath(v, path, t)
}

func (t *IPAddrList) RemovePath(v any, path []string) error {
	return SliceRemovePath(v, path, t)
}

var (
	_ yaml.Marshaler   = IPPrefixList{}
	_ yaml.Unmarshaler = &IPPrefixList{}
	_ Reader           = IPPrefixList{}
	_ Writer           = &IPPrefixList{}
	_ PathReader       = IPPrefixList{}
	_ PathWriter       = &IPPrefixList{}
	_ PathAdder        = &IPPrefixList{}
	_ PathRemover      = &IPPrefixList{}
)

func (v IPPrefixList) MarshalYAML() (any, error) {
	return formatCommaSeperated(v), nil
}

func (t *IPPrefixList) UnmarshalYAML(n *yaml.Node) error {
	s := ""
	err := n.Decode(&s)
	if err != nil {
		return err
	}

	err = writeCommaSeperated(s, t)
	return err
}

func (v IPPrefixList) Read(t any) (err error) {
	switch t := t.(type) {
	case *string:
		*t = formatCommaSeperated(v)
	default:
		err = ReadSlice(v, t)
	}
	return
}

func (t *IPPrefixList) Write(v any) (err error) {
	switch v := v.(type) {
	case string:
		err = writeCommaSeperated(v, t)
	default:
		err = WriteSlice(v, t)
	}
	return
}

func (v IPPrefixList) ReadPath(path []string, t any) error {
	return ReadSlice(v, t)
}

func (t *IPPrefixList) WritePath(v any, path []string) error {
	return WriteSlice(v, t)
}

func (t *IPPrefixList) AddPath(v any, path []string) error {
	return SliceAddPath(v, path, t)
}

func (t *IPPrefixList) RemovePath(v any, path []string) error {
	return SliceRemovePath(v, path, t)
}

type IPAddr netip.Addr

var (
	_ yaml.Marshaler   = IPAddr{}
	_ yaml.Unmarshaler = &IPAddr{}
	_ Reader           = IPAddr{}
	_ Writer           = &IPAddr{}
)

func (v IPAddr) MarshalYAML() (any, error) {
	addr := netip.Addr(v)
	if !addr.IsValid() {
		return nil, nil
	}
	return addr.String(), nil
}

func (t *IPAddr) UnmarshalYAML(n *yaml.Node) (err error) {
	s := ""
	err = n.Decode(&s)
	if err != nil {
		return err
	}
	addr, err := netip.ParseAddr(s)
	*t = IPAddr(addr)
	return
}

func (v IPAddr) Read(t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *netip.Addr:
		*t = netip.Addr(v)
	case *IPAddr:
		*t = v
	case *string:
		*t = netip.Addr(v).String()
	default:
		err = errRead(v, t)
	}
	return
}

func (t *IPAddr) Write(v any) (err error) {
	switch v := v.(type) {
	case netip.Addr:
		*t = IPAddr(v)
	case IPAddr:
		*t = v
	case string:
		var addr netip.Addr
		addr, err = netip.ParseAddr(v)
		*t = IPAddr(addr)
	default:
		err = errWrite(v, *t)
	}
	return
}

func (v IPAddr) String() string {
	return netip.Addr(v).String()
}

type IPPrefix netip.Prefix

func (v *IPPrefix) Get() any {
	return *v
}

var (
	_ yaml.Marshaler   = IPAddr{}
	_ yaml.Unmarshaler = &IPAddr{}
	_ Reader           = IPAddr{}
	_ Writer           = &IPAddr{}
)

func (v IPPrefix) MarshalYAML() (any, error) {
	prefix := netip.Prefix(v)
	if !prefix.IsValid() {
		return nil, nil
	}
	return prefix.String(), nil
}

func (t *IPPrefix) UnmarshalYAML(n *yaml.Node) (err error) {
	s := ""
	err = n.Decode(&s)
	if err != nil {
		return err
	}
	addr, err := netip.ParsePrefix(s)
	*t = IPPrefix(addr)
	return
}

func (v IPPrefix) Read(t any) (err error) {
	switch t := t.(type) {
	case *any:
		*t = v
	case *netip.Prefix:
		*t = netip.Prefix(v)
	case *IPPrefix:
		*t = v
	case *string:
		*t = netip.Prefix(v).String()
	default:
		err = errRead(v, t)
	}
	return
}

func (t *IPPrefix) Write(v any) (err error) {
	switch v := v.(type) {
	case netip.Prefix:
		*t = IPPrefix(v)
	case IPPrefix:
		*t = v
	case string:
		var prefix netip.Prefix
		prefix, err = netip.ParsePrefix(v)
		*t = IPPrefix(prefix)
	default:
		err = errWrite(v, *t)
	}
	return
}

func (v IPPrefix) String() string {
	return netip.Prefix(v).String()
}
