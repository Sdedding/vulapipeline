//go:build linux

package organize

import (
	"fmt"
	"strings"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func ShowPeerConfig(c *core.PeerConfig) *StyledOutput {
	o := &StyledOutput{}

	o.AddKeyValue(Bold("peer"), Yellow(toBase64String(c.PublicKey)))
	o.AddNewline()
	if len(c.PresharedKey) > 0 {
		o.AddKeyValue(Text("preshared key"), Text("(hidden)"))
	}

	if !c.EndpointAddr.IsUnspecified() {
		o.AddKeyValue(Text("endpoint"), Text(fmt.Sprintf("%s:%d", c.EndpointAddr, c.EndpointPort)))
	}

	if len(c.AllowedIPs) > 0 {
		allowedIPs := make([]string, len(c.AllowedIPs))
		for i, ip := range c.AllowedIPs {
			allowedIPs[i] = ip.String()
		}
		o.AddKeyValue(Text("allowed ips"), Text(strings.Join(allowedIPs, ", ")))
	} else {
		o.AddKeyValue(Text("allowed ips"), Text("(none)"))
	}

	if c.LatestHandshake != (time.Time{}) {
		o.AddKeyValue(Text("latest handshake"), Text(fmt.Sprintf("%s ago", time.Since(c.LatestHandshake))))
	}

	if c.Stats.RxBytes != 0 || c.Stats.TxBytes != 0 {
		o.AddKeyValue(Text("transfer"), Text(fmt.Sprintf("%d received, %d sent", c.Stats.RxBytes, c.Stats.TxBytes)))
	}

	if c.PersistentKeepalive != 0 {
		o.AddKeyValue(Text("persistent keepalive"), Text(fmt.Sprintf("every %d seconds", c.PersistentKeepalive)))
	}

	return o
}
