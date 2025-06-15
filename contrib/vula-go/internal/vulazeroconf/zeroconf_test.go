package vulazeroconf

import (
	"encoding/base64"
	"log"
	"strings"
	"testing"
)

func TestParseEscaped(t *testing.T) {
	// arrange
	input, err := base64.StdEncoding.DecodeString("cD1cMjUzXDI1NVwyNTVcMjU1XDI1NVwyMjNcMjE3XDIyOVwyNTBKXDAzMFwxMzJcMTM4XDE2MVwwMTB6")
	if err != nil {
		log.Fatal(err)
	}

	// act
	result, err := parseEscaped(string(input))
	if err != nil {
		t.Fatal(err)
	}

	// assert
	if len(result) != 18 {
		t.Errorf("len(result) = %d", len(result))
	}
	if !strings.HasPrefix(string(result), "p=") {
		t.Errorf("%s missing prefix p=", string(result[0]))
	}
}
