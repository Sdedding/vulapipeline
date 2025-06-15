package vulazeroconf

import (
	"fmt"
	"net/netip"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

const (
	zeroconfServiceType = "_opabinia._udp"
	zeroconfDomain      = "local."
)

func serializeDescriptorZeroconf(d *schema.Descriptor) (text []string) {
	items := []schema.StringMapItem{}

	err := schema.Read(d, nil, &items)
	if err != nil {
		panic(err)
	}

	// python uses a wier format, we have to follow
	v4a := make([]byte, 0, len(d.V4A)*4)
	for _, addr := range d.V4A {
		v4a = append(v4a, netip.Addr(addr).AsSlice()...)
	}
	v6a := make([]byte, 0, len(d.V6A)*16)
	for _, addr := range d.V6A {
		v6a = append(v6a, netip.Addr(addr).AsSlice()...)
	}
	p := netip.Addr(d.P).AsSlice()

	for _, item := range items {
		s := []byte{}
		s = append(s, item.Key...)
		s = append(s, '=')
		switch item.Key {
		case "v4a":
			s = append(s, v4a...)
		case "v6a":
			s = append(s, v6a...)
		case "p":
			s = append(s, p...)
		default:
			s = append(s, item.Value...)
		}

		text = append(text, string(s))
	}
	return
}

func parseDescriptorZeroconf(text []string) (d *schema.Descriptor, err error) {
	// python vula decided to format v4a and v6a as byte-sequences
	// to conform to this, we just convert the received string to the corresponding ip addr

	items := []schema.StringMapItem{}
	for _, line := range text {
		sep := strings.IndexRune(line, '=')
		if sep < 0 {
			err = fmt.Errorf("missing field separator")
			return
		}
		key := line[:sep]
		value := line[sep+1:]

		switch key {
		case "v4a":
			var addrs []string
			var addrBytes []byte
			addrBytes, err = parseEscaped(value)
			if err != nil {
				return
			}
			addrs, err = parseAddrsFromBytes(addrBytes, 4)
			if err != nil {
				return
			}
			value = strings.Join(addrs, ",")
		case "v6a":
			var addrs []string
			var addrBytes []byte
			addrBytes, err = parseEscaped(value)
			if err != nil {
				return
			}
			addrs, err = parseAddrsFromBytes(addrBytes, 16)
			if err != nil {
				return
			}
			value = strings.Join(addrs, ",")
		case "p":
			if len(value) > 0 {
				// for some reason zeroconf does a very nasty encoding for p
				var addrBytes []byte
				addrBytes, err = parseEscaped(value)
				if err != nil {
					return
				}
				addr, ok := netip.AddrFromSlice(addrBytes)
				if !ok {
					err = fmt.Errorf("invalid addr: %s", value)
					return
				}
				value = addr.String()
			}
		}

		items = append(items, schema.StringMapItem{Key: key, Value: value})
	}

	d = &schema.Descriptor{}
	err = schema.Write(items, nil, d)
	return
}

func parseAddrsFromBytes(v []byte, length int) (addrs []string, err error) {
	if len(v)%length != 0 {
		fmt.Println(len(v), v)
		err = fmt.Errorf("the number of bytes must be a multiple of %d", length)
		return
	}
	for i := 0; i < len(v); i += length {
		addr, _ := netip.AddrFromSlice(v[i : i+length])
		addrs = append(addrs, addr.String())
	}
	return
}

func parseEscaped(s string) (b []byte, err error) {
	status := 0
	n := 0
	var ad int

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch status {
		case 0:
			if c == '\\' {
				status = 1
			} else {
				b = append(b, c)
			}
		case 1:
			ad, err = readAsciiDigit(c)
			if err != nil {
				return
			}
			n = ad * 100
			status = 2
		case 2:
			ad, err = readAsciiDigit(c)
			if err != nil {
				return
			}
			n += ad * 10
			status = 3
		case 3:
			ad, err = readAsciiDigit(c)
			if err != nil {
				return
			}
			n += ad
			status = 0
			if n > 255 {
				err = fmt.Errorf("escape sequence out of range: %d", n)
				return
			}
			b = append(b, byte(n))
		}
	}

	if status != 0 {
		err = fmt.Errorf("open escape sequence")
	}
	return
}

func readAsciiDigit(d byte) (i int, err error) {
	if d < '0' || d > '9' {
		err = fmt.Errorf("no ascii digit: %c", d)
	}
	i = int(d - '0')
	return
}
