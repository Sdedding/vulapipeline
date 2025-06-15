package organize

import (
	"encoding/base64"
	"fmt"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func ShowPeer(p *core.Peer, stats *core.PeerStats) *StyledOutput {
	o := &StyledOutput{}
	// header
	showPeer(p, o)
	o.AddNewline()
	// id
	o.AddIndentation()
	o.AddKeyValue(Text("id"), Text(peerID(p)))
	o.AddNewline()
	// status
	o.AddIndentation()
	showPeerStatus(p, o)
	o.AddNewline()
	// endpoint
	o.AddIndentation()
	o.AddKeyValue(Text("endpoint"), Text(peerEndpoint(p)))
	o.AddNewline()
	// allowed ips
	o.AddIndentation()
	o.AddKeyValue(Text("allowed ips"), JoinStyledStringer(", ", peerWgAllowedIPs(p)...))
	o.AddNewline()
	// disabled ips
	disabledIPs := peerEnabledIPs(p, false)
	if len(disabledIPs) > 0 {
		o.AddIndentation()
		o.AddKeyValue(Text("disabled ips"), JoinStyledStringer(", ", disabledIPs...))
		o.AddNewline()
	}
	// latest signature
	o.AddIndentation()
	showPeerLatestSignature(p, o)
	o.AddNewline()
	// latest handshake
	o.AddIndentation()
	showPeerLatestHandshake(stats, o)
	o.AddNewline()
	// transfer
	o.AddIndentation()
	showPeerTransfer(stats, o)
	o.AddNewline()
	// wg pubkey
	o.AddIndentation()
	o.AddKeyValue(Text("wg pubkey"), Text(base64.StdEncoding.EncodeToString(p.Descriptor.WireGuardPK)))
	// other names
	otherNames := peerOtherNames(p)
	if len(otherNames) > 0 {
		o.AddNewline()
		o.AddIndentation()
		o.AddKeyValue(Text("other names"), JoinStyledString(", ", otherNames...))
	}
	return o
}

func showPeer(p *core.Peer, o *StyledOutput) {
	var greenOrYellow Color
	if p.Pinned && p.Verified {
		greenOrYellow = GreenColor
	} else {
		greenOrYellow = YellowColor
	}

	o.AddKeyValue(Colored("peer", greenOrYellow), Colored(peerName(p), greenOrYellow))
}

/*
func showPeerWarning(stats *core.PeerStats, o *StyledOutput) {
	warning := ""
	if stats == nil {
		warning = "wireguard peer is not configured"
	}

	o.AddKeyValue(Red("warning"), Red(warning))
}
*/

func showPeerStatus(p *core.Peer, o *StyledOutput) {
	items := make([]StyledString, 0, 4)

	if p.Enabled {
		items = append(items, Green("enabled"))
	} else {
		items = append(items, Red("disabled"))
	}

	if p.Pinned {
		items = append(items, Green("pinned"))
	} else {
		items = append(items, Yellow("unpinned"))
	}

	if p.Verified {
		items = append(items, Green("verified"))
	} else {
		color := DefaultColor
		if p.Pinned {
			color = RedColor
		} else {
			color = YellowColor
		}
		items = append(items, Colored("unverified", color))
	}

	if p.UseAsGateway {
		items = append(items, StyledString{
			Text:   "gateway",
			Color:  BlueColor,
			Weight: BoldWeight,
		})
	}

	o.AddKeyValue(Text("status"), JoinStyled(" ", items...)...)
}

func showPeerLatestSignature(p *core.Peer, o *StyledOutput) {
	deltaSeconds := time.Now().Unix() - p.Descriptor.ValidStart
	delta := time.Second * time.Duration(deltaSeconds)
	o.AddKeyValue(Text("latest signature"), Text(fmt.Sprintf("%s ago", delta)))
}

func showPeerLatestHandshake(stats *core.PeerStats, o *StyledOutput) {
	var value StyledString
	if stats == nil || !stats.HasLatestHandshake {
		value = Yellow("none")
	} else {
		deltaSeconds := time.Now().Unix() - stats.LatestHandshake
		delta := time.Second * time.Duration(deltaSeconds)
		value = Text(fmt.Sprintf("%s ago", delta))
	}

	o.AddKeyValue(Text("latest handshake"), value)
}

func showPeerTransfer(stats *core.PeerStats, o *StyledOutput) {
	if stats == nil || stats.LatestHandshake+stats.TxBytes+stats.RxBytes == 0 {
		o.AddKeyValue(Text("transfer"), Text("0"))
	}

	o.AddKeyValue(Text("transfer"), Text(fmt.Sprintf("%d received, %d sent", stats.RxBytes, stats.TxBytes)))
}
