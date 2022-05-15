package blockchain

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewMiner(t *testing.T) {
	fmt.Println("TestNewClient:")
	// Create a fake net
	net := NewFakeNet()

	//
	// Generate some keys
	privKey1, pubKey1, _ := GenerateKeypair()
	privKey2, pubKey2, _ := GenerateKeypair()
	privKey3, pubKey3, _ := GenerateKeypair()
	privKey4, pubKey4, _ := GenerateKeypair()
	privKey5, pubKey5, _ := GenerateKeypair()
	privKey6, pubKey6, _ := GenerateKeypair()
	address1 := GenerateAddress(pubKey1)
	address2 := GenerateAddress(pubKey2)
	address3 := GenerateAddress(pubKey3)
	address4 := GenerateAddress(pubKey4)
	address5 := GenerateAddress(pubKey5)
	address6 := GenerateAddress(pubKey6)
	newBalances := make(map[string]uint32)
	newBalances[address1] = 233
	newBalances[address2] = 99
	newBalances[address3] = 67
	newBalances[address4] = 400
	newBalances[address5] = 300

	genesis, config, _ := MakeGenesisDefault(newBalances)

	client1 := NewClient("Alice", net, genesis, privKey1)
	client2 := NewClient("Bob", net, genesis, privKey2)
	client3 := NewClient("Cindy", net, genesis, privKey3)
	miner1 := NewMiner("Minnie", net, NUM_ROUNDS_MINING, genesis, privKey4)
	miner2 := NewMiner("Mickey", net, NUM_ROUNDS_MINING, genesis, privKey5)
	miner3 := NewMiner("Donald", net, NUM_ROUNDS_MINING, genesis, privKey6)

	showBalancesC := func(c *Client) {
		fmt.Printf("Alice has %d gold\n", c.LastBlock.BalanceOf(client1.GetAddress()))
		fmt.Printf("Bob has %d gold\n", c.LastBlock.BalanceOf(client2.GetAddress()))
		fmt.Printf("Cindy has %d gold\n", c.LastBlock.BalanceOf(client3.GetAddress()))
		fmt.Printf("Minnie has %d gold\n", c.LastBlock.BalanceOf(miner1.GetAddress()))
		fmt.Printf("Mickey has %d gold\n", c.LastBlock.BalanceOf(miner2.GetAddress()))
		fmt.Printf("Donald has %d gold\n", c.LastBlock.BalanceOf(miner3.GetAddress()))
	}

	showBalancesM := func(m *Miner) {
		fmt.Printf("Alice has %d gold\n", m.LastBlock.BalanceOf(client1.GetAddress()))
		fmt.Printf("Bob has %d gold\n", m.LastBlock.BalanceOf(client2.GetAddress()))
		fmt.Printf("Cindy has %d gold\n", m.LastBlock.BalanceOf(client3.GetAddress()))
		fmt.Printf("Minnie has %d gold\n", m.LastBlock.BalanceOf(miner1.GetAddress()))
		fmt.Printf("Mickey has %d gold\n", m.LastBlock.BalanceOf(miner2.GetAddress()))
		fmt.Printf("Donald has %d gold\n", m.LastBlock.BalanceOf(miner3.GetAddress()))
	}

	net.Register(client1, client2, client3, miner1, miner2)

	if client1.GetAddress() != address1 || client2.GetAddress() != address2 || client3.GetAddress() != address3 {
		t.Fatalf("TestNewMiner: Error, Incorrect client address")
	}

	if miner1.GetAddress() != address4 || miner2.GetAddress() != address5 || miner3.GetAddress() != address6 {
		t.Fatal("TestNewMiner: Error, Inconnect miner address")
	}

	miner1.Initialize()
	miner2.Initialize()

	// A new transaction from client1 to client2
	output1 := Output{Address: client2.GetAddress(), Amount: 200}
	outputs := []Output{output1}

	client1.PostTransaction(outputs, config.defaultTxFee)

	go func() {
		time.Sleep(2 * time.Second)
		net.Register(miner3)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println()
		fmt.Printf("Minnie has a chain of length %d\n", miner1.CurrentBlock.ChainLength)

		fmt.Println()
		fmt.Printf("Mickey has a chain of length %d\n", miner2.CurrentBlock.ChainLength)

		fmt.Println()
		fmt.Printf("Donald has a chain of length %d\n", miner3.CurrentBlock.ChainLength)

		fmt.Println()
		fmt.Println("Final Balances (Minnie's perspective):")
		showBalancesM(miner1)

		fmt.Println()
		fmt.Println("Final Balances (Alice's perspective):")
		showBalancesC(client1)

		fmt.Println()
		fmt.Println("Final Balances (Donald's perspective):")
		showBalancesM(miner3)

		os.Exit(0)
	}()

}
