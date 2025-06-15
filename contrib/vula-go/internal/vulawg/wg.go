package vulawg

import "codeberg.org/vula/vula/contrib/vula-go/internal/core"

type Interface struct {
	deviceName string
}

var _ core.WireguardInterface = &Interface{}

func NewInterface(deviceName string) *Interface {
	return &Interface{deviceName}
}

func (w *Interface) Name() string {
	return w.deviceName
}

func (w *Interface) Configuration() (privateKey, publicKey []byte, listenPort uint16, firewallMark int, err error) {
	return getConfiguration(w.deviceName)
}

func (w *Interface) SetConfiguration(privateKey []byte, listenPort uint16, firewallMark int) error {
	return setConfiguration(w.deviceName, privateKey, int(listenPort), firewallMark)
}

func (w *Interface) Peers() ([]core.PeerConfig, error) {
	return getPeers(w.deviceName)
}

func (w *Interface) SetPeer(config *core.PeerConfig, dryRun bool) ([]string, error) {
	log, err := applyPeerConfig(w.deviceName, config, dryRun)
	for _, line := range log {
		core.LogDebug(line)
	}
	return log, err
}
