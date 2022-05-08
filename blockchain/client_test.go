package blockchain

import (
	"fmt"
	"testing"
)

func TestNewClient(t *testing.T) {
	fmt.Println("TestNewClient:")
	// Create a fake net
	net := NewFakeNet()

	//
	// Generate some keys
	privKey1, pubKey1, _ := GenerateKeypair()
	privKey2, pubKey2, _ := GenerateKeypair()
	privKey3, pubKey3, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)
	address3 := GenerateAddress(pubKey3)
	newBalances := map[string]uint32{address1: 1000, address2: 500}
	genesis, config, _ := MakeGenesisDefault(newBalances)

	client1 := NewClient("Alice", net, genesis, privKey1)
	client2 := NewClient("Bob", net, genesis, privKey2)
	client3 := NewClient("Cindy", net, genesis, privKey3)

	net.Register(client1, client2, client3)

	if client1.GetAddress() != address1 || client2.GetAddress() != address2 || client3.GetAddress() != address3 {
		t.Fatalf("TestNewClient: Error, Incorrect client address")
	}

	// Second block
	target := CalculateTarget(POW_LEADING_ZEROES)
	secondBlock := NewBlock(address3, genesis, target, config.coinbaseAmount)

	if secondBlock == nil {
		t.Fatalf(`Failed to make the second block`)
	}

	// A new transaction from client1 to client2
	var nonce uint32 = 0
	output1 := Output{Address: address2, Amount: 200}
	outputs := []Output{output1}

	tx, err1 := NewTransaction(address1, nonce, pubKey1, nil, config.defaultTxFee, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	signature := tx.Sign(privKey1)
	if signature == nil {
		t.Fatalf("Sign() Error: Failed to sign transaction")
	}

	secondBlock.AddTransaction(tx)

	// find proof
	secondBlock.Rerun(genesis)

	var MaxUint uint32 = ^uint32(0)
	var verified bool = false
	for i := uint32(0); i <= MaxUint; i++ {
		(*secondBlock).Proof = i
		if secondBlock.hasValidProof() {
			verified = true
			break
		}
	}

	if verified == false {
		t.Fatalf("Failed to find proof\n")
	}

	// ReceiveBlock
	client3.ReceiveBlock(*secondBlock)
	client2.ReceiveBlock(*secondBlock)
	client2.ShowBlockchain()
	client3.ShowBlockchain()

	thirdBlock := NewBlock(address3, secondBlock, target, config.coinbaseAmount)
	if thirdBlock == nil {
		t.Fatalf(`Failed to make the third block`)
	}

	thirdBlock.Rerun(secondBlock)

	verified = false
	for i := uint32(0); i <= MaxUint; i++ {
		(*thirdBlock).Proof = i
		if thirdBlock.hasValidProof() {
			verified = true
			break
		}
	}
	if verified == false {
		t.Fatalf("Failed to find proof\n")
	}

	client3.ReceiveBlock(*thirdBlock)
	client2.ReceiveBlock(*thirdBlock)
	client1.ReceiveBlock(*thirdBlock)
	client1.ShowBlockchain()

}
