package blockchain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSet(t *testing.T) {
	// Generate Keys
	rng := rand.Reader
	privKey, _ := rsa.GenerateKey(rng, 2048)
	pubKey := &(*privKey).PublicKey

	// Create Address from pubKey
	pubKeyBytes := sha256.Sum256([]byte(fmt.Sprintf("%x||%x", (*pubKey).N, (*pubKey).E)))
	from := hex.EncodeToString(pubKeyBytes[:])

	// nonce and outputs, sig and data will be nil for now.
	var nonce uint32 = 0
	output1 := Output{Address: "address1", Amount: 100}
	output2 := Output{Address: "address2", Amount: 200}
	outputs := []Output{output1, output2}

	tx, _ := NewTransaction(from, nonce, pubKey, nil, 100, outputs, nil)
	tx.Sign(privKey)

	s := NewSet[*Transaction]()
	s.Add(tx)

	if !s.Contains(tx) {
		t.Fatalf("Error: set fails to add item to Set")
	}

	// Generate Keys
	privKey1, pubKey1, _ := GenerateKeypair()

	// Create Address from pubKey
	from1 := GenerateAddress(pubKey1)
	tx1, _ := NewTransaction(from1, nonce, pubKey, nil, 100, outputs, nil)

	tx1.Sign(privKey1)

	if s.Contains(tx1) {
		t.Fatalf("Error: set contains non exist item")
	}

	s.Add(tx1)

	if s.Size() != 2 {
		t.Fatalf("Error: set does not have 2 items")
	}

	arr := s.ToArray()
	for _, v := range arr {
		fmt.Printf("set contains " + v.GetHashStr() + "\n")
	}

	s.Remove(tx)

	if s.Contains(tx) || !s.Contains(tx1) {
		t.Fatalf("Error: set fails to delete a item")
	}

}
