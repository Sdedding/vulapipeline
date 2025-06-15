package filerepo

import (
	"fmt"
	"os"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"codeberg.org/vula/vula/contrib/vula-go/internal/schema"
	"gopkg.in/yaml.v3"
)

type KeyFile struct {
	FileName string
}

var _ core.KeyRepository = &KeyFile{}

func (r *KeyFile) Read() (*core.Keys, error) {
	err := fixKeyFilePermission(r.FileName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(r.FileName)
	if err != nil {
		return nil, err
	}

	content := schema.KeyFile{}
	err = yaml.Unmarshal(data, &content)
	if err != nil {
		return nil, err
	}

	return importKeys(&content), nil
}

func (r *KeyFile) Write(k *core.Keys) error {
	s := exportKeys(k)
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	err = os.WriteFile(r.FileName, data, 0o600)
	return err
}

func (r *KeyFile) MoveBackup() error {
	nowUnix := time.Now().UTC().Unix()
	return os.Rename(r.FileName, fmt.Sprintf("%s_.bad.%d", r.FileName, nowUnix))
}

func fixKeyFilePermission(fileName string) error {
	return os.Chmod(fileName, 0o600)
}

func importKeys(c *schema.KeyFile) *core.Keys {
	// TODO: validate
	keys := &core.Keys{
		PqCtidhP512: core.Keypair{
			SK: c.PqCtidhP512SecKey, // len: 74
			PK: c.PqCtidhP512PubKey, // len: 64
		},
		VkEd25519: core.Keypair{
			SK: c.VkEd25519SecKey, // len: 32
			PK: c.VkEd25519PubKey, // len: 32
		},
		WgCurve25519: core.Keypair{
			SK: c.WgCurve25519SecKey, // len: 32
			PK: c.WgCurve25519PubKey, // len: 32
		},
	}

	return keys
}

func exportKeys(k *core.Keys) *schema.KeyFile {
	return &schema.KeyFile{
		PqCtidhP512SecKey:  k.PqCtidhP512.SK,
		PqCtidhP512PubKey:  k.PqCtidhP512.PK,
		VkEd25519SecKey:    k.VkEd25519.SK,
		VkEd25519PubKey:    k.VkEd25519.PK,
		WgCurve25519SecKey: k.WgCurve25519.SK,
		WgCurve25519PubKey: k.WgCurve25519.PK,
	}
}
