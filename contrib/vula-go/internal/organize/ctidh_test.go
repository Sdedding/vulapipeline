package organize

import (
	"bytes"
	"encoding/base64"
	"testing"

	"codeberg.org/vula/highctidh/src/ctidh512"
)

func TestHkdfOrganize(t *testing.T) {
	type testCase struct {
		Secret   string
		Expected string
	}

	testCases := []testCase{
		{"my_raw_key", "Y52eWgiYuPYtHlnqZpRqAG2USxILzRS57s61ePUdWO4="},
		{"test string", "P39kOvTABj0XVj0wFMcZZw1F/njgFOlJDE44i8QG2LA="},
	}

	for _, testCase := range testCases {
		t.Logf("testing secret %s", testCase.Secret)
		s, err := hkdfOrganize([]byte(testCase.Secret))
		if err != nil {
			t.Fatal(err)
		}

		b := base64.StdEncoding.EncodeToString(s)
		if b != testCase.Expected {
			t.Errorf("s = %s", s)
		}
	}

}

func TestCtidh(t *testing.T) {
	// arrange
	pk, err := base64.StdEncoding.DecodeString("Z3KMO4sMCRKAHwIfEz5Iu0LvTOdqyLZWFsCIuMxhNWIYpFM8fZgvS8tebrqyLsBBUgnS2g1NvxCH1XrEMdRpTQ==")
	if err != nil {
		t.Fatal(err)
	}
	sk, err := base64.StdEncoding.DecodeString("/Ab9AAUDAgUFAgcB+fv++v0B+QAAAfsDAP4D+wX9BAAB/AH9AP4C/QD9BgD9AQAAAfj7Af8D/f8CAgL6/P8BAAABAv4B///7AP8=")
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ctidh512.NewEmptyPrivateKey()
	err = privateKey.FromBytes(sk)
	if err != nil {
		t.Fatal(err)
	}

	expectedRawKey, err := base64.StdEncoding.DecodeString("JFrTxJfJEATgHy18J6WtLbEEtCxSQuJxAsRMgxgUG5+t4LaRTCfLBCuXyMp4oz7VJ1K5wmJlsv3o88MAnsbYRg==")
	if err != nil {
		t.Fatal(err)
	}

	expectedPsk, err := base64.StdEncoding.DecodeString("Uhuga5CX8JZASkXfNMwG2Eqm3kHUsYgGb5thJe+4F28=")
	if err != nil {
		t.Fatal(err)
	}

	sut := newCtidh512Impl(privateKey)

	// act
	rawKey := sut.DH(pk)
	if err != nil {
		t.Fatal(err)
	}

	psk, err := hkdfOrganize(rawKey)
	if err != nil {
		t.Fatal(err)
	}

	// assert
	if !bytes.Equal(expectedRawKey, rawKey) {
		t.Errorf("rawKey is %s", base64.StdEncoding.EncodeToString(rawKey))
	}

	if !bytes.Equal(psk, expectedPsk) {
		t.Errorf("psk is %s", base64.StdEncoding.EncodeToString(psk))
	}
}
