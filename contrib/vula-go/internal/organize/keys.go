package organize

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
	"golang.org/x/crypto/nacl/sign"
)

func hkdfOrganize(rawKey []byte) ([]byte, error) {
	key := hkdf.Extract(sha512.New, rawKey, nil)
	r := hkdf.Expand(sha512.New, key, []byte("vula-organize-1"))
	s := make([]byte, 32)
	_, err := io.ReadFull(r, s)
	return s, err
}

func curve25519KeypairGen() (*core.Keypair, error) {
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &core.Keypair{
		SK: key.Bytes(),
		PK: key.PublicKey().Bytes(),
	}, nil
}

func ed25519KeypairGen() (*core.Keypair, error) {
	pk, sk, err := sign.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	return &core.Keypair{
		SK: sk[:],
		PK: pk[:],
	}, nil
}

func ctidhKeypairGen() *core.Keypair {
	return generateCtidhKeypair()
}

func genKeys() (*core.Keys, error) {
	vkEd25519, err := ed25519KeypairGen()
	if err != nil {
		return nil, err
	}

	wgCurve25519, err := curve25519KeypairGen()
	if err != nil {
		return nil, err
	}

	return &core.Keys{
		PqCtidhP512:  *ctidhKeypairGen(),
		VkEd25519:    *vkEd25519,
		WgCurve25519: *wgCurve25519,
	}, nil
}

func generateOrReadKeys(r core.KeyRepository) (keys *core.Keys, err error) {
	keys, err = r.Read()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			core.LogInfo("Keys file not found")
		} else {
			err = r.MoveBackup()
			if err != nil {
				return
			}
		}

		keys, err = genKeys()
		if err != nil {
			return
		}

		err = r.Write(keys)
	}
	return
}
