package core

import "time"

type Systemd interface {
	GetUnit(name string) (SystemdUnit, error)
}

type SystemdUnit interface {
	GetActiveState() (string, error)
	GetStateChangeTimeStamp() (time.Time, error)
}

type DbusMeta interface {
	NameHasOwner(name string) (bool, error)
}
