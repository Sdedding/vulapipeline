package core

type Keys struct {
	PqCtidhP512  Keypair
	VkEd25519    Keypair
	WgCurve25519 Keypair
}

type Keypair struct {
	SK []byte
	PK []byte
}

type KeyRepository interface {
	Read() (*Keys, error)
	Write(keys *Keys) error
	MoveBackup() error
}
