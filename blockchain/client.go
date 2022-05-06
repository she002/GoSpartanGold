package blockchain

import (
	"crypto/rsa"
	"fmt"

	"github.com/chuckpreslar/emission"
)

// instead of using a constructor, we use a struct here to represent a client
type Client struct {
	Name                        string
	Address                     string
	privKey                     *rsa.PrivateKey
	pubKey                      *rsa.PublicKey
	Blocks                      map[string]*Block
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	PendingBlocks               map[string][]*Block
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceiveBlock                *Block
	Config                      BlockchainConfig
	Nonce                       uint32
	Net                         *FakeNet
	Emitter                     *emission.Emitter
}

type Message struct {
	Address       string
	PrevBlockHash string
}

// The genesis block can only be set if the client does not already have the genesis block.
func (c *Client) setGenesisBlock(startingBlock *Block) {
	if c.LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain.")
	}
	c.LastConfirmedBlock = startingBlock
	c.LastBlock = startingBlock
	blockId, _ := startingBlock.GetHash()
	c.Blocks[blockId] = startingBlock
}

// The amount of gold available to the client looking at the last confirmed block
func (c *Client) confirmedBalance() uint32 {
	return c.LastConfirmedBlock.BalanceOf(c.Address)
}

// Any gold received in the last confirmed block or before
func (c *Client) availableGold() uint32 {
	var pendingSpent uint32 = 0
	for _, tx := range c.PendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return c.confirmedBalance() - pendingSpent
}

// Broadcasts a transaction from the client giving gold to the clients
func (c *Client) postTransaction(outputs []Output, fee uint32) *Transaction {
	if fee < 0 {
		fee = c.Config.defaultTxFee
	}
	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > c.availableGold() {
		// modify here
		panic(`Account doesn't have enough balance for transaction`)
	}
	// add data to the constructor
	tx, _ := NewTransaction(c.Address, c.Nonce, c.pubKey, nil, fee, outputs, nil)

	tx.Sign(c.privKey)
	c.PendingOutgoingTransactions[tx.Id()] = tx
	c.Nonce++
	c.Net.broadcast(POST_TRANSACTION, tx)

	return tx
}

// Validates and adds a block to the list of blocks, possibly
// updating the head of the blockchain.
func (c *Client) receiveBlock(b *Block, bs []byte) *Block {
	block := b
	var err error = nil
	if b == nil {
		fmt.Println("Deseralize")
		block, err = BytesToBlock(bs)
		if err != nil {
			panic("Failed to deseralize block")
		}
	}
	blockId, _ := block.GetHash()
	if _, received := c.Blocks[blockId]; received {
		return nil
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		fmt.Printf("Block %v does not have a valid proof", blockId)
		return nil
	}

	var prevBlock *Block = nil
	prevBlock, received := c.Blocks[block.PrevBlockHash]
	if !received {
		if !prevBlock.IsGenesisBlock() {
			stuckBlocks, received := c.PendingBlocks[block.PrevBlockHash]
			if !received {
				c.requestMissingBlock(block)
				// TODO: Define a set
				stuckBlocks = make([]*Block, 10)
			}
			stuckBlocks = append(stuckBlocks, block)
			c.PendingBlocks[block.PrevBlockHash] = stuckBlocks
			return nil
		}
	}

	if !block.IsGenesisBlock() {
		if !block.Rerun(prevBlock) {
			return nil
		}
	}
	blockId, _ = block.GetHash()
	c.Blocks[blockId] = block

	if c.LastBlock.ChainLength < block.ChainLength {
		c.LastBlock = block
		c.setLastConfirmed()
	}
	// TODO: Review code on client.js
	unstuckBlocks := make([]*Block, 0)

	blockId, _ = block.GetHash()
	if val, received := c.PendingBlocks[blockId]; received {
		unstuckBlocks = val
	}

	delete(c.PendingBlocks, blockId)

	for _, uBlock := range unstuckBlocks {
		fmt.Printf("processing unstuck block %v", blockId)
		// Need to change the "" into empty []byte
		c.receiveBlock(uBlock, nil)
	}
	return block
}

// Request the previous block from the network.
// convert []byte into string
func (c *Client) requestMissingBlock(block *Block) {
	fmt.Printf("Asking for missing block: %v", block.PrevBlockHash)
	var msg = Message{c.Address, block.PrevBlockHash}
	c.Net.broadcast(MISSING_BLOCK, msg)
}

// Resend any transactions in the pending list
func (c *Client) resendPendingTransactions() {
	for _, tx := range c.PendingOutgoingTransactions {
		c.Net.broadcast(POST_TRANSACTION, tx)
	}
}

// Takes an object representing a request for a missing block
func (c *Client) provideMissingBlock(msg Message) {
	if val, received := c.Blocks[msg.PrevBlockHash]; received {
		fmt.Printf("Providing missing block %v", msg.PrevBlockHash)
		block := val
		c.Net.sendMessage(msg.Address, PROOF_FOUND, block)
	}
}

// Sets the last confirmed block according to the most accepted block and also
// updating pending transactions according to this block.
func (c *Client) setLastConfirmed() {
	block := c.LastBlock
	confirmedBlockHeight := block.ChainLength - CONFIRMED_DEPTH
	if confirmedBlockHeight < 0 {
		confirmedBlockHeight = 0
	}
	for block.ChainLength > confirmedBlockHeight {
		block, _ = c.Blocks[block.PrevBlockHash]
	}
	c.LastConfirmedBlock = block
	for id, tx := range c.PendingOutgoingTransactions {
		if c.LastConfirmedBlock.contains(tx) {
			delete(c.PendingOutgoingTransactions, id)
		}
	}
}

// Utility method that displays all confirmed balances for all clients
func (c *Client) showAllBalances() {
	fmt.Printf("Showing balances:")
	for id, balance := range c.LastConfirmedBlock.Balances {
		fmt.Printf("	%v", id)
		fmt.Printf("	%v", balance)
		fmt.Println("")
	}
}

// Logs messages to stdout
func (c *Client) log(msg Message) {
	name := c.Address[0:10]
	if len(c.Name) > 0 {
		name = c.Name
	}
	fmt.Printf("	%v", name)
	fmt.Printf("	%v", msg)
}

// Print out the blocks in the blockchain from the current head to the genesis block.
func (c *Client) showBlockchain() {
	block := c.LastBlock
	fmt.Println("BLOCKCHAIN:")
	for block != nil {
		blockId, _ := block.GetHash()
		fmt.Println(blockId)
		block = c.Blocks[block.PrevBlockHash]
	}

}

func NewClient(name string, Net *FakeNet, startingBlock *Block, keyPair *rsa.PrivateKey) *Client {
	var c Client
	c.Net = Net
	c.Name = name

	if keyPair == nil {
		c.privKey, c.pubKey, _ = GenerateKeypair()
	} else {
		c.privKey = keyPair
		c.pubKey = &keyPair.PublicKey
	}
	c.Address = GenerateAddress(c.pubKey)
	c.Nonce = 0

	c.PendingOutgoingTransactions = make(map[string]*Transaction)
	c.PendingReceivedTransactions = make(map[string]*Transaction)
	c.Blocks = make(map[string]*Block)
	c.PendingBlocks = make(map[string][]*Block)

	if startingBlock != nil {
		c.setGenesisBlock(startingBlock)
	}

	receive := func(b *Block) {
		// TODO
		c.receiveBlock(b, nil)
	}

	c.Emitter = emission.NewEmitter()
	c.Emitter.On(PROOF_FOUND, receive)
	c.Emitter.On(MISSING_BLOCK, c.provideMissingBlock)
	return &c
}
