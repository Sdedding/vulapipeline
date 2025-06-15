package core

import "net/netip"

type WireguardSystem interface {
	// SyncWireguardInterface gets or creates a wireguard interface
	SyncInterface(name string, primaryIP netip.Addr, dryRun bool) ([]string, error)

	// GetWireguardInterface gets access to the specified wireguard interface
	GetInterface(name string) WireguardInterface
}

type WireguardInterface interface {
	// Name gest the name of the interface
	Name() string

	// Configuration gets the current interface configuration
	Configuration() (privateKey, publicKey []byte, listenPort uint16, firewallMark int, err error)

	// SetConfiguration configures the wireguard interface
	SetConfiguration(privateKey []byte, listenPort uint16, firewallMark int) error

	// Peers gets all peer configurations
	Peers() ([]PeerConfig, error)

	// SetPeer configures the specified peer
	SetPeer(config *PeerConfig, dryRun bool) ([]string, error)
}
