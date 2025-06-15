package organize

import (
	"bytes"
	"slices"
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"golang.org/x/crypto/nacl/sign"
)

func buildDescriptorSignBuf(d *schema.Descriptor) []byte {
	buf := []byte{}
	var items []schema.StringMapItem

	// extract the fields from the descriptor
	err := schema.Read(d, nil, &items)
	if err != nil {
		panic("descriptor should be convertable to map[string]string")
	}

	slices.SortFunc(items, func(a, b schema.StringMapItem) int {
		return strings.Compare(a.Key, b.Key)
	})

	for _, item := range items {
		if item.Key == "s" {
			continue
		}

		if len(buf) > 0 {
			buf = append(buf, ' ')
		}

		buf = append(buf, item.Key...)
		buf = append(buf, '=')
		buf = append(buf, item.Value...)
		buf = append(buf, ';')
	}
	return buf
}

// signDescriptor adds a signature to the descriptor
func signDescriptor(d *schema.Descriptor, seed []byte) error {
	_, privKey, err := sign.GenerateKey(bytes.NewReader(seed))
	if err != nil {
		return err
	}

	bufToSign := buildDescriptorSignBuf(d)
	outBuf := make([]byte, 0, len(bufToSign)+sign.Overhead)
	sign.Sign(outBuf, bufToSign, privKey)
	d.S = outBuf[:sign.Overhead]
	return nil
}

// verifyDescriptorSignature  Verifies the signature. Returns true if valid, false if invalid.
func verifyDescriptorSignature(d *schema.Descriptor) bool {
	if len(d.VK) != 32 || len(d.S) != sign.Overhead {
		return false
	}
	vk := (*[32]byte)(d.VK)
	bufToSign := buildDescriptorSignBuf(d)
	signed := make([]byte, len(bufToSign)+sign.Overhead)
	copy(signed, d.S)
	copy(signed[sign.Overhead:], bufToSign)

	_, valid := sign.Open(bufToSign, signed, vk)
	return valid
}
