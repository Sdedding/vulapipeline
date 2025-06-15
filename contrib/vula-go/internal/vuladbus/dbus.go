package vuladbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	organizeDbusName = "local.vula.organize"
	discoverDbusName = "local.vula.discover"
	publishDbusName  = "local.vula.publish"

	organizeDbusPath = "/local/vula/organize"
	discoverDbusPath = "/local/vula/discover"
	publishDbusPath  = "/local/vula/publish"
)

const (
	// discoverDbusListeName is the name of the vula.discover1.Listen DBUS interface
	discoverDbusListenName = "local.vula.discover1.Listen"
	publishDbusListenName  = "local.vula.publish1.Listen"
)

// dbusIntrospectableName is the name of the DBUS introspectable interface
const dbusIntrospectableName = "org.freedesktop.DBus.Introspectable"

type ErrDbusNameTaken string

func (e ErrDbusNameTaken) Error() string {
	return fmt.Sprintf("vula: DBUS name %s is already taken", string(e))
}

func acquireDbusName(conn *dbus.Conn, name string) error {
	reply, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		return ErrDbusNameTaken(name)
	}

	return nil
}

type MetaClient struct {
	conn       *dbus.Conn
	dbusObject dbus.BusObject
}

func (d *MetaClient) NameHasOwner(name string) (bool, error) {
	call := d.dbusObject.Call("org.freedesktop.DBus.NameHasOwner", 0, name)
	var response bool
	if call.Err != nil {
		return false, call.Err
	}
	err := call.Store(&response)
	if err != nil {
		return false, err
	}
	return response, nil
}

func NewMetaClient() (*MetaClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	systemd := conn.Object(
		"org.freedesktop.DBus",
		"/org/freedesktop/DBus",
	)
	return &MetaClient{
		conn:       conn,
		dbusObject: systemd,
	}, nil
}
