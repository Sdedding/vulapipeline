package vuladbus

import (
	"strings"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

// discoverListenIntrospectable is the introspectable of the discover1.Listen DBUS interface
var discoverListenIntrospectable = introspect.NewIntrospectable(&introspect.Node{
	Interfaces: []introspect.Interface{
		introspect.IntrospectData,
		{
			Name: discoverDbusListenName,
			Methods: []introspect.Method{
				{
					Name: "listen",
					Args: []introspect.Arg{
						{Name: "ip_addrs", Type: "as", Direction: "in"},
						{Name: "our_wg_pk", Type: "s", Direction: "in"},
					},
				},
			},
		},
	},
})

func RunDiscoverServer(done <-chan struct{}, listen core.DiscoverListen) error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer util.CloseLog(conn)

	err = exportDbusDiscoverListen(conn, &discoverListen{listen})
	if err != nil {
		return err
	}

	<-done
	return nil
}

type DiscoverClient struct {
	conn           *dbus.Conn
	discoverObject dbus.BusObject
}

func NewDiscoverClient() (*DiscoverClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	discoverObject := conn.Object(
		discoverDbusName,
		discoverDbusPath,
	)
	return &DiscoverClient{
		conn:           conn,
		discoverObject: discoverObject,
	}, nil
}

func (d *DiscoverClient) Close() error {
	return d.conn.Close()
}

func (d *DiscoverClient) Listen(addrs []string, ourWgPK string) error {
	return d.discoverObject.Call(discoverDbusListenName+".listen", 0, addrs, ourWgPK).Err
}

// discoverListen implements vula.discover1.Listen DBUS interface
type discoverListen struct {
	d core.DiscoverListen
}

// TODO: own wg pk here is unused
// Listen implements vula1.discover1.Listen.listen DBUS interface method
func (l *discoverListen) Listen(ipAddrs []string, ourWgPK string) *dbus.Error {
	core.LogDebugf("local.vula.discover1.Listen called with ip addresses: %s", strings.Join(ipAddrs, ", "))
	err := l.d.Listen(ipAddrs, "")
	if err != nil {
		return dbus.MakeFailedError(err)
	}
	return nil
}

func exportDbusDiscoverListen(conn *dbus.Conn, listen *discoverListen) error {
	err := conn.ExportWithMap(listen, map[string]string{"Listen": "listen"}, discoverDbusPath, discoverDbusListenName)
	if err != nil {
		return err
	}
	err = conn.Export(discoverListenIntrospectable, discoverDbusPath, dbusIntrospectableName)
	if err != nil {
		return err
	}

	err = acquireDbusName(conn, discoverDbusName)
	return err
}
