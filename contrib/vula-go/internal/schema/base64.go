package schema

import (
	"encoding/base64"

	"gopkg.in/yaml.v3"
)

type Base64 []byte

var (
	_ yaml.Unmarshaler = &Base64{}
	_ yaml.Marshaler   = Base64{}
	_ Reader           = Base64{}
	_ Writer           = &Base64{}
)

func (b *Base64) UnmarshalYAML(n *yaml.Node) error {
	s := ""
	err := n.Decode(&s)
	if err != nil {
		return err
	}

	*b, err = base64.StdEncoding.DecodeString(s)
	return err
}

func (b Base64) MarshalYAML() (any, error) {
	return base64.StdEncoding.EncodeToString(b), nil
}

func (b Base64) Read(t any) (err error) {
	switch t := t.(type) {
	case *Base64:
		*t = b
	case *[]byte:
		*t = b
	case *string:
		*t = base64.StdEncoding.EncodeToString(b)
	default:
		err = errRead(b, t)
	}
	return
}

func (b *Base64) Write(v any) (err error) {
	switch v := v.(type) {
	case Base64:
		*b = v
	case []byte:
		*b = v
	case string:
		*b, err = base64.StdEncoding.DecodeString(v)
	default:
		err = errWrite(v, *b)
	}
	return
}

func (v Base64) String() string {
	return base64.StdEncoding.EncodeToString(v)
}
