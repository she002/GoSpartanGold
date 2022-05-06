package blockchain

import (
	"fmt"
	"testing"
)

func TestGenesisBlock(t *testing.T) {
	// Generate some keys
	_, pubKey1, _ := GenerateKeypair()
	_, pubKey2, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)

	newBalances := map[string]uint32{address1: 100, address2: 100}

	genesis, _, _ := MakeGenesisDefault(newBalances)
	if genesis == nil {
		t.Fatalf(`Failed to make a genesis block`)
	}

	if genesis.IsGenesisBlock() == false {
		t.Fatalf(`Failed to make a genesis block`)
	}

	fmt.Println(genesis.ToString())
}

func TestNewBlock(t *testing.T) {
	// Generate some keys
	_, pubKey1, _ := GenerateKeypair()
	_, pubKey2, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)

	newBalances := map[string]uint32{address1: 100, address2: 100}

	genesis, config, _ := MakeGenesisDefault(newBalances)
	if genesis == nil {
		t.Fatalf(`Failed to make a genesis block`)
	}

	if genesis.IsGenesisBlock() == false {
		t.Fatalf(`Failed to make a genesis block`)
	}

	target := CalculateTarget(POW_LEADING_ZEROES)
	secondBlock := NewBlock("", genesis, target, config.coinbaseAmount)

	if secondBlock == nil {
		t.Fatalf(`Failed to make the second block`)
	}
	fmt.Println(secondBlock.ToString())
}

func TestAddTransction(t *testing.T) {

	// Generate some keys
	privKey1, pubKey1, _ := GenerateKeypair()
	_, pubKey2, _ := GenerateKeypair()
	_, pubKey3, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)
	address3 := GenerateAddress(pubKey3)

	newBalances := map[string]uint32{address1: 1000, address2: 1000, address3: 1000}

	genesis, config, _ := MakeGenesisDefault(newBalances)
	if genesis == nil {
		t.Fatalf(`Failed to make a genesis block`)
	}

	if genesis.IsGenesisBlock() == false {
		t.Fatalf(`Failed to make a genesis block`)
	}

	target := CalculateTarget(POW_LEADING_ZEROES)
	secondBlock := NewBlock("", genesis, target, config.coinbaseAmount)

	if secondBlock == nil {
		t.Fatalf(`Failed to make the second block`)
	}

	// nonce and outputs, sig and data will be nil for now.
	var nonce uint32 = 0
	output2 := Output{Address: address2, Amount: 200}
	output3 := Output{Address: address3, Amount: 300}
	outputs := []Output{output2, output3}

	tx, err1 := NewTransaction(address1, nonce, pubKey1, nil, config.defaultTxFee, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	signature := tx.Sign(privKey1)
	if signature == nil {
		t.Fatalf("Sign() Error: Failed to sign transaction")
	}

	secondBlock.AddTransaction(tx)

	fmt.Println(secondBlock.ToString())
}

func TestVerifyProof(t *testing.T) {
	// Generate some keys
	privKey1, pubKey1, _ := GenerateKeypair()
	_, pubKey2, _ := GenerateKeypair()
	_, pubKey3, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)
	address3 := GenerateAddress(pubKey3)

	newBalances := map[string]uint32{address1: 1000, address2: 1000, address3: 1000}

	genesis, config, _ := MakeGenesisDefault(newBalances)
	if genesis == nil {
		t.Fatalf(`Failed to make a genesis block`)
	}

	if genesis.IsGenesisBlock() == false {
		t.Fatalf(`Failed to make a genesis block`)
	}

	target := CalculateTarget(POW_LEADING_ZEROES)
	secondBlock := NewBlock("", genesis, target, config.coinbaseAmount)

	if secondBlock == nil {
		t.Fatalf(`Failed to make the second block`)
	}

	// nonce and outputs, sig and data will be nil for now.
	var nonce uint32 = 0
	output2 := Output{Address: address2, Amount: 200}
	output3 := Output{Address: address3, Amount: 300}
	outputs := []Output{output2, output3}

	tx, err1 := NewTransaction(address1, nonce, pubKey1, nil, config.defaultTxFee, outputs, nil)
	if err1 != nil {
		t.Fatalf(`NewTransaction(from, nonce, pubkey, outputs) Error: %v`, err1)
	}

	signature := tx.Sign(privKey1)
	if signature == nil {
		t.Fatalf("Sign() Error: Failed to sign transaction")
	}

	secondBlock.AddTransaction(tx)

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
	} else {
		fmt.Printf("Proof is %d\n", secondBlock.Proof)
	}

}
