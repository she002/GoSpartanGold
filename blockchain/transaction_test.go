package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewTransaction(t *testing.T) {

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

	msg, err := NewTransaction(from, nonce, pubKey, nil, 100, outputs, nil)
	if err != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err)
	}
	fmt.Println((*msg).ToString())
}

func TestToBytes(t *testing.T) {
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

	msg, err1 := NewTransaction(from, nonce, pubKey, nil, 100, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	data, err2 := TransactionToBytes(msg)
	if err2 != nil {
		t.Fatalf(`TransactionToBytes() Error: %v`, err2)
	}

	tran1, err3 := BytesToTransaction(data)
	if err3 != nil {
		t.Fatalf(`BytesToTransaction() Error: %v`, err3)
	}
	fmt.Println(tran1.ToString())
}

func TestSign(t *testing.T) {

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

	tx, err1 := NewTransaction(from, nonce, pubKey, nil, 100, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	signature := tx.Sign(privKey)
	if signature == nil {
		t.Fatalf("Sign() Error: Failed to sign transaction")
	}

	result := tx.VerifySignature()
	if result != true {
		t.Fatalf("VerifySignature() Error: failed to verify signature")
	}
	fmt.Println("signture verified")

}

func TestTotalOutput(t *testing.T) {
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

	tx, err1 := NewTransaction(from, nonce, pubKey, nil, 100, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	value := tx.TotalOutput()
	if value != 300 {
		t.Fatalf(`TotalOutput() Error: expect to return 300, but actually return %d`, value)
	}
	fmt.Printf("Total output is %d\n", value)
}
