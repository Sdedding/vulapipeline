package organize

import (
	"crypto/sha3"
	"encoding/base64"
	"io"
	"sync"

	"codeberg.org/vula/highctidh/src/ctidh512"
	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

func generateCtidhKeypair() *core.Keypair {
	sk, pk := ctidh512.GenerateKeyPair()
	return &core.Keypair{
		SK: sk.Bytes(),
		PK: pk.Bytes(),
	}
}

type ctidh512Impl struct {
	mux        sync.Mutex
	cache      map[string][]byte
	privateKey *ctidh512.PrivateKey
}

func newCtidh512Impl(privateKey *ctidh512.PrivateKey) *ctidh512Impl {
	return &ctidh512Impl{
		cache:      map[string][]byte{},
		privateKey: privateKey,
	}
}

func (c *ctidh512Impl) DH(pk []byte) (rawKey []byte) {
	core.LogDebugf("Generating CTIDH PSK for pk %s", toBase64String(pk))
	publicKey := ctidh512.NewPublicKey(pk)
	rawKey = ctidh512.DeriveSecret(c.privateKey, publicKey)

	xof := sha3.NewSHAKE256()
	_, err := xof.Write(rawKey)
	if err != nil {
		panic(err)
	}
	_, err = io.ReadFull(xof, rawKey)
	if err != nil {
		panic(err)
	}

	return rawKey
}

func (c *ctidh512Impl) CachedDH(pk []byte) (rawKey []byte) {
	c.mux.Lock()
	defer c.mux.Unlock()

	pkString := base64.StdEncoding.EncodeToString(pk)
	ok := false
	rawKey, ok = c.cache[pkString]
	if ok {
		return
	}

	rawKey = c.DH(pk)
	c.cache[pkString] = rawKey
	return
}

func (c *ctidh512Impl) GetPSK(pk []byte) ([]byte, error) {
	rawKey := c.CachedDH(pk)
	return hkdfOrganize(rawKey)
}
