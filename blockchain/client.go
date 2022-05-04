package blockchain

import (
	"crypto/rsa"
	"fmt"

	// import "./utils"
	"github.com/chuckpreslar/emission"
)

// instead of using a constructor, we use a struct here to represent a client
// Upper case first letter to change the scope to global
type Client struct {
	Name                        string
	Address                     string
	KeyPair                     *rsa.PrivateKey
	Blocks                      map[string]*Block
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	PendingBlocks               map[string][]*Block
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceiveBlock                *Block
	BlockChain                  *BlockChain
	Nonce                       int
	Net                         *FakeNet          `json:"_"`
	Emitter                     *emission.Emitter `json:"_"`
}

type Message struct {
	Address       string
	PrevBlockHash []byte
}

// The genesis block can only be set if the client does not already have the genesis block.
func (c *Client) setGenesisBlock(startingBlock *Block) {
	if c.LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain.")
	}
	c.LastConfirmedBlock = startingBlock
	c.LastBlock = startingBlock
	c.Blocks[string(startingBlock.id())] = startingBlock
}

// The amount of gold available to the client looking at the last confirmed block
func (c Client) confirmedBalance() int {
	return c.LastConfirmedBlock.balanceOf(c.Address)
}

// Any gold received in the last confirmed block or before
func (c Client) availableGold() int {
	var pendingSpent int = 0
	for _, tx := range c.PendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return int(c.confirmedBalance()) - pendingSpent
}

// Broadcasts a transaction from the client giving gold to the clients
func (c *Client) postTransaction(outputs []Output, fee int) *Transaction {
	if fee < 0 {
		fee = c.BlockChain.DefaultTxFee
	}
	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > c.availableGold() {
		panic(`Account doesn't have enough balance for transaction`)
	}

	tx := NewTransaction(c.Address, c.Nonce, &c.KeyPair.PublicKey, nil, fee, outputs)

	tx.Sign(c.KeyPair)
	c.PendingOutgoingTransactions[string(tx.Id())] = tx
	c.Nonce++
	c.Net.broadcast(POST_TRANSACTION, tx)

	return tx
}

// Validates and adds a block to the list of blocks, possibly
// updating the head of the blockchain.
func (c *Client) receiveBlock(b *Block, bs string) *Block {
	block := b
	if b == nil {
		fmt.Println("Deseralize")
		block = c.BlockChain.deserializeBlock([]byte(bs))
	}
	if _, received := c.Blocks[string(block.id())]; received {
		return nil
	}
	if !block.hasValidProof() && !block.IsGenesisBlock() {
		fmt.Printf("Block %v does not have a valid proof", string(block.id()))
		return nil
	}

	var prevBlock *Block = nil
	prevBlock, received := c.Blocks[string(block.PrevBlockHash)]
	if !received {
		if !prevBlock.IsGenesisBlock() {
			stuckBlocks, received := c.PendingBlocks[string(block.PrevBlockHash)]
			if !received {
				c.requestMissingBlock(*block)
				stuckBlocks = make([]*Block, 10)
			}
			stuckBlocks = append(stuckBlocks, block)
			c.PendingBlocks[string(block.PrevBlockHash)] = stuckBlocks
			return nil
		}
	}

	if !block.IsGenesisBlock() {
		if !block.rerun(prevBlock) {
			return nil
		}
	}

	c.Blocks[string(block.id())] = block

	if c.LastBlock.ChainLength < block.ChainLength {
		c.LastBlock = block
		c.setLastConfirmed()
	}

	unstuckBlocks := make([]*Block, 0)
	if val, received := c.PendingBlocks[string(block.id())]; received {
		unstuckBlocks = val
	}

	delete(c.PendingBlocks, string(block.id()))

	for _, uBlock := range unstuckBlocks {
		fmt.Printf("processing unstuck block %v", string(block.id()))
		c.receiveBlock(uBlock, "")
	}
	return block
}

// Request the previous block from the network.
func (c Client) requestMissingBlock(block Block) {
	fmt.Printf("Asking for missing block: %v", block.PrevBlockHash)
	var msg = Message{c.Address, block.PrevBlockHash}
	c.Net.broadcast(MISSING_BLOCK, msg)
}

// Resend any transactions in the pending list
func (c Client) resendPendingTransactions() {
	for _, tx := range c.PendingOutgoingTransactions {
		c.Net.broadcast(POST_TRANSACTION, tx)
	}
}

// Takes an object representing a request for a missing block
func (c Client) provideMissingBlock(msg Message) {
	if val, received := c.Blocks[string(msg.PrevBlockHash)]; received {
		fmt.Printf("Providing missing block %v", string(msg.PrevBlockHash))
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
		block = c.Blocks[string(block.PrevBlockHash)]
	}
	c.LastConfirmedBlock = block
	for id, tx := range c.PendingOutgoingTransactions {
		if c.LastConfirmedBlock.contains(tx) {
			delete(c.PendingOutgoingTransactions, id)
		}
	}
}

// Utility method that displays all confirmed balances for all clients
func (c Client) showAllBalances() {
	fmt.Printf("Showing balances:")
	for id, balance := range c.LastConfirmedBlock.Balances {
		fmt.Printf("	%v", id)
		fmt.Printf("	%v", balance)
		fmt.Println("")
	}
}

// Logs messages to stdout
func (c Client) log(msg Message) {
	name := c.Address[0:10]
	if len(c.Name) > 0 {
		name = c.Name
	}
	fmt.Printf("	%v", name)
	fmt.Printf("	%v", msg)
}

// Print out the blocks in the blockchain from the current head to the genesis block.
func (c Client) showBlockchain() {
	block := c.LastBlock
	fmt.Println("BLOCKCHAIN:")
	for block != nil {
		fmt.Println(block.id())
		block = c.Blocks[string(block.PrevBlockHash)]
	}

}

func NewClient(name string, Net *FakeNet, startingBlock *Block, keyPair *rsa.PrivateKey) *Client {
	var c Client
	c.Net = Net
	c.Name = name

	if keyPair == nil {
		c.KeyPair = utils.GenerateKeypair()
	}
	else {
		c.KeyPair = keyPair
	}
	c.Address = utils.CalculateAddress(&c.KeyPair.PublicKey)
	c.Nonce = 0

	c.PendingOutgoingTransactions = make(map[string]*Transaction)
	c.PendingReceivedTransactions = make(map[string]*Transaction)
	c.Blocks = make(map[string]*Block)
	c.PendingBlocks = make(map[string][]*Block)
	
	if startingBlock != nil {
		c.setGenesisBlock(startingBlock)
	}

	receive := func(b *Block) {
		c.receiveBlock(b, "")
	}

	c.Emitter = emission.NewEmitter()
	c.Emitter.On(PROOF_FOUND, receive)
	c.Emitter.On(MISSING_BLOCK, c.provideMissingBlock)
	return &c
}
