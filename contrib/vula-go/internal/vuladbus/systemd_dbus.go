package vuladbus

import (
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"github.com/godbus/dbus/v5"
)

type SystemdClientUnit struct {
	unitObject dbus.BusObject
}

var _ core.SystemdUnit = &SystemdClientUnit{}

func (d *SystemdClientUnit) GetActiveState() (string, error) {
	call := d.unitObject.Call(
		"org.freedesktop.DBus.Properties.Get",
		0,
		"org.freedesktop.systemd1.Unit",
		"ActiveState",
	)
	if call.Err != nil {
		return "", call.Err
	}
	response := ""
	err := call.Store(&response)
	return response, err
}

func (d *SystemdClientUnit) GetStateChangeTimeStamp() (time.Time, error) {
	call := d.unitObject.Call(
		"org.freedesktop.DBus.Properties.Get",
		0,
		"org.freedesktop.systemd1.Unit",
		"StateChangeTimestamp",
	)
	if call.Err != nil {
		return time.Time{}, call.Err
	}
	var posixTimestamp int64
	err := call.Store(&posixTimestamp)
	timeStamp := time.UnixMicro(posixTimestamp)
	return timeStamp, err
}

type SystemdClient struct {
	conn          *dbus.Conn
	systemdObject dbus.BusObject
}

var _ core.Systemd = &SystemdClient{}

func NewSystemdClient() (*SystemdClient, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	systemd := conn.Object(
		"org.freedesktop.systemd1",
		"/org/freedesktop/systemd1",
	)
	return &SystemdClient{
		conn:          conn,
		systemdObject: systemd,
	}, nil
}

func (d *SystemdClient) GetUnit(name string) (core.SystemdUnit, error) {
	call := d.systemdObject.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, name)
	if call.Err != nil {
		return nil, call.Err
	}
	var unitPath dbus.ObjectPath
	err := call.Store(&unitPath)
	if err != nil {
		return nil, err
	}

	unitObject := d.conn.Object("org.freedesktop.systemd1", unitPath)
	return &SystemdClientUnit{
		unitObject: unitObject,
	}, nil
}
