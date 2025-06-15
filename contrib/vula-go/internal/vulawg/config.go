package vulawg

import (
	"fmt"
	"math"

	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func getConfiguration(interfaceName string) (privateKey, publicKey []byte, listenPort uint16, firewallMark int, err error) {
	client, err := wgctrl.New()
	if err != nil {
		return
	}
	defer util.CloseLog(client)

	device, err := client.Device(interfaceName)
	if err != nil {
		return
	}

	privateKey = device.PrivateKey[:]
	publicKey = device.PublicKey[:]
	if device.ListenPort < 0 || device.ListenPort > math.MaxUint16 {
		return nil, nil, 0, 0, fmt.Errorf("ListenPort must be between 0 and 65535")
	}
	listenPort = uint16(device.ListenPort)
	firewallMark = device.FirewallMark
	return
}

func setConfiguration(deviceName string, privateKey []byte, listenPort int, firewallMark int) error {
	wgPrivateKey, err := wgtypes.NewKey(privateKey)
	if err != nil {
		return err
	}

	config := wgtypes.Config{
		PrivateKey:   &wgPrivateKey,
		ListenPort:   &listenPort,
		FirewallMark: &firewallMark,
	}

	client, err := wgctrl.New()
	if err != nil {
		return err
	}
	defer util.CloseLog(client)

	return client.ConfigureDevice(deviceName, config)
}
