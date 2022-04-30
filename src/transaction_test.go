package blockchain

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
)

func TestNewTransaction(t *testing.T) {

	from := "random address"
	var nonce uint32 = 123
	pubkey := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	output1 := Output{Address: "address1", Amount: 100}
	output2 := Output{Address: "address2", Amount: 200}
	outputs := []Output{output1, output2}

	msg, err := NewTransaction(from, nonce, pubkey, outputs)
	if err != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err)
	}
	fmt.Println((*msg).ToString())
}

func TestToBytes(t *testing.T) {
	from := "random address"
	var nonce uint32 = 123
	pubkey := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	output1 := Output{Address: "address1", Amount: 100}
	output2 := Output{Address: "address2", Amount: 200}
	outputs := []Output{output1, output2}

	msg, err1 := NewTransaction(from, nonce, pubkey, outputs)
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
	from := "random address"
	var nonce uint32 = 123
	pubkey := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	output1 := Output{Address: "address1", Amount: 100}
	output2 := Output{Address: "address2", Amount: 200}
	outputs := []Output{output1, output2}

	tx, err1 := NewTransaction(from, nonce, pubkey, outputs)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}
	rng := rand.Reader
	privKey, err2 := rsa.GenerateKey(rng, 2048)
	if err2 != nil {
		t.Fatalf("rsa.GenerateKey(random io.Reader, bits int)) Error: Failed to generate  private Key")
	}

	signature := tx.Sign(privKey)
	if signature == nil {
		t.Fatalf("Sign() Error: Failed to sign transaction")
	}

	result := tx.VerifySignature(privKey, signature)
	if result != true {
		t.Fatalf("VerifySignature() Error: failed to verify signature")
	}
	fmt.Println("signture verified")

}

func TestTotalOutput(t *testing.T) {
	from := "random address"
	var nonce uint32 = 123
	pubkey := [32]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	output1 := Output{Address: "address1", Amount: 100}
	output2 := Output{Address: "address2", Amount: 200}
	outputs := []Output{output1, output2}

	tx, err1 := NewTransaction(from, nonce, pubkey, outputs)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	value := tx.TotalOutput()
	if value != 300 {
		t.Fatalf(`TotalOutput() Error: expect to return 300, but actually return %d`, value)
	}
	fmt.Printf("Total output is %d\n", value)
}
