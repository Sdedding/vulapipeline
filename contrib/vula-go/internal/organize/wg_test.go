package organize

import (
	"strings"
	"testing"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func TestPresharedKeyHidden(t *testing.T) {
	// arrange
	c := &core.PeerConfig{PresharedKey: make([]byte, 42)}
	needle := "preshared key\x1b[0m: \x1b[0m(hidden)"

	// act
	output := ShowPeerConfig(c)
	text := string(output.ToConsole())

	// assert
	ok := strings.Contains(text, needle)
	if !ok {
		t.Errorf("text should contain '%s'", needle)
	}

}
