package organize

import (
	"bytes"
	"encoding/base64"
	"math"
	"net/netip"
	"slices"
	"testing"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
)

func TestParseDescriptorString(t *testing.T) {
	// arrange
	c, err := base64.StdEncoding.DecodeString("NnoGEZ4W+d6TE22+Qyau0LF513FM43EagOP9aiSX9KhTCS1Gryt7qDoM04j7p0KQRJxwkcPEO/MpIJE5/bJKYQ==")
	if err != nil {
		t.Fatal(err)
	}
	pk, err := base64.StdEncoding.DecodeString("3w5/xje5jsdUCX30JfS/L/bMuwZRniK69dAVprN7t3c=")
	if err != nil {
		t.Fatal(err)
	}
	v6a := schema.IPAddrList{
		schema.IPAddr(netip.MustParseAddr("fdff:ffff:ffdf:989f:24cf:bda:1262:cfc6")),
		schema.IPAddr(netip.MustParseAddr("fe80::bc92:4dff:fe82:30d")),
		schema.IPAddr(netip.MustParseAddr("fd54:f27a:17c1:3a61::2")),
	}
	vk, err := base64.StdEncoding.DecodeString("afToKyN29ubu4DkhUMLoGIt5WjbsgEHYuccNtxvbjmA=")
	if err != nil {
		t.Fatal(vk)
	}

	// act
	d, err := schema.ParseDescriptorString(testDescriptorUnsigned)
	if err != nil {
		t.Fatal(err)
	}

	// assert
	if !bytes.Equal(d.C, c) {
		t.Errorf("d.C = %s", toBase64String(c))
	}
	if d.DT != 86400 {
		t.Errorf("d.DT = %d", d.DT)
	}
	if d.E {
		t.Errorf("d.E = %v", d.E)
	}
	if d.Hostname != "vula-bookworm-test1.local." {
		t.Errorf("d.Hostname = %s", d.Hostname)
	}
	if d.P.String() != "fdff::1" {
		t.Errorf("d.P = %s", d.P)
	}
	if !bytes.Equal(d.PK, pk) {
		t.Errorf("d.PK = %s", toBase64String(d.PK))
	}
	if d.Port != 5354 {
		t.Errorf("d.Port = %d", d.Port)
	}
	if len(d.R) != 0 {
		t.Errorf("d.R = %s", d.R)
	}
	if len(d.V4A) != 1 || d.V4A[0].String() != "10.89.0.2" {
		t.Errorf("d.V4A = %v", d.V4A)
	}
	if !slices.Equal(d.V6A, v6a) {
		t.Errorf("d.V6A = %s", v6a)
	}
	if d.VF != 1743974365 {
		t.Errorf("d.VF = %d", d.VF)
	}
	if !bytes.Equal(d.VK, vk) {
		t.Errorf("d.VK = %s", toBase64String(d.VK))
	}
}

func TestSerializeDescriptor(t *testing.T) {
	// arrange
	d, err := schema.ParseDescriptorString(testDescriptorUnsigned)
	if err != nil {
		t.Fatal(err)
	}
	expectedString := "c=NnoGEZ4W+d6TE22+Qyau0LF513FM43EagOP9aiSX9KhTCS1Gryt7qDoM04j7p0KQRJxwkcPEO/MpIJE5/bJKYQ==; dt=86400; e=0; hostname=vula-bookworm-test1.local.; p=fdff::1; pk=3w5/xje5jsdUCX30JfS/L/bMuwZRniK69dAVprN7t3c=; port=5354; r=; v4a=10.89.0.2; v6a=fdff:ffff:ffdf:989f:24cf:bda:1262:cfc6,fe80::bc92:4dff:fe82:30d,fd54:f27a:17c1:3a61::2; vf=1743974365; vk=afToKyN29ubu4DkhUMLoGIt5WjbsgEHYuccNtxvbjmA=;"

	// act
	s := schema.SerializeDescriptorString(&d)

	// assert
	if s != expectedString {
		t.Errorf("s = %s", s)
	}
}

func TestCheckDescriptorFreshness(t *testing.T) {
	// arrange
	type testCase struct {
		validStart    int64
		validDuration int64
		now           time.Time
		expected      bool
	}

	testCases := []testCase{
		{1749924000, 10, time.Unix(1749924000, 0), true},
		{1749924000, 10, time.Unix(1749924011, 0), false},
		{math.MaxInt64 - 1, math.MaxInt64 - 1, time.Unix(math.MaxInt64-1, 0), false},
	}

	for i, tc := range testCases {
		d := &core.Descriptor{ValidStart: tc.validStart, ValidDuration: tc.validDuration}

		// act
		isFresh := checkDescriptorFreshness(d, tc.now)

		// assert
		if isFresh != tc.expected {
			t.Errorf("test %d: isFresh = %v", i, isFresh)
		}
	}
}
