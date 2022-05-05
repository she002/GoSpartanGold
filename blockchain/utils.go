package blockchain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const DEFAULT_RSA_KEYLENGTH int = 2048

func GenerateKeypair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	rng := rand.Reader
	privKey, err := rsa.GenerateKey(rng, DEFAULT_RSA_KEYLENGTH)
	if err != nil {
		return nil, nil, err
	}
	pubKey := &(*privKey).PublicKey
	return privKey, pubKey, nil
}

func GenerateAddress(pubKey *rsa.PublicKey) string {
	pubKeyBytes := sha256.Sum256([]byte(fmt.Sprintf("%x||%x", (*pubKey).N, (*pubKey).E)))
	from := hex.EncodeToString(pubKeyBytes[:])
	return from
}
