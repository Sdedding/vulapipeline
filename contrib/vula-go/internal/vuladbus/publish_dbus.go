package vuladbus

import (
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/util"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

var publishListenName = "local.vula.publish1.Listen"

var publishListenIntrospectable = introspect.NewIntrospectable(&introspect.Node{
	Interfaces: []introspect.Interface{
		introspect.IntrospectData,
		{
			Name: publishListenName,
			Methods: []introspect.Method{
				{
					Name: "listen",
					Args: []introspect.Arg{
						{Name: "new_announcements", Type: "a{ss}", Direction: "in"},
					},
				},
			},
		},
	},
})

type publishListen struct {
	p core.PublishListen
}

func (l *publishListen) Listen(newAnnouncements map[string]string) *dbus.Error {
	core.LogDebugf("received new announcements: %v", newAnnouncements)
	err := l.p.Listen(newAnnouncements)
	if err != nil {
		return dbus.MakeFailedError(err)
	}
	return nil
}

func RunPublishServer(done <-chan struct{}, listen core.PublishListen) error {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer util.CloseLog(conn)

	err = exportDbusPublishListen(conn, &publishListen{listen})
	if err != nil {
		return err
	}

	<-done
	return nil
}

type PublishClient struct {
	conn          *dbus.Conn
	publishObject dbus.BusObject
}

func NewPublishClient() (*PublishClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	publishObject := conn.Object(
		publishDbusName,
		publishDbusPath,
	)
	return &PublishClient{
		conn:          conn,
		publishObject: publishObject,
	}, nil
}

func (d *PublishClient) Close() error {
	return d.conn.Close()
}

func (d *PublishClient) Listen(newAnnouncements map[string]string) error {
	return d.publishObject.Call(publishDbusListenName+".listen", 0, newAnnouncements).Err
}

func exportDbusPublishListen(conn *dbus.Conn, listen *publishListen) error {
	err := conn.ExportWithMap(listen, map[string]string{"Listen": "listen"}, publishDbusPath, publishListenName)
	if err != nil {
		return err
	}
	err = conn.Export(publishListenIntrospectable, publishDbusPath, dbusIntrospectableName)
	if err != nil {
		return err
	}

	err = acquireDbusName(conn, publishDbusName)
	return err
}
