package blockchain

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chuckpreslar/emission"
)

// instead of using a constructor, we use a struct here to represent a client
type Client struct {
	Name                        string
	Address                     string
	PrivKey                     *rsa.PrivateKey
	PubKey                      *rsa.PublicKey
	Blocks                      map[string]*Block
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	PendingBlocks               map[string]*Set[*Block]
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceivedBlock               *Block
	Config                      BlockchainConfig
	Nonce                       uint32
	Net                         *FakeNet
	Emitter                     *emission.Emitter
	mu                          sync.Mutex
}

type Message struct {
	Address       string
	PrevBlockHash string
}

// The genesis block can only be set if the client does not already have the genesis block.
func (c *Client) SetGenesisBlock(startingBlock *Block) {

	if (*c).LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain.")
	}
	(*c).LastConfirmedBlock = startingBlock
	(*c).LastBlock = startingBlock
	blockId, err := startingBlock.GetHash()
	if err != nil {
		panic("Failed to get block hash")
	}
	(*c).Blocks[blockId] = startingBlock
}

// The amount of gold available to the client looking at the last confirmed block
func (c *Client) ConfirmedBalance() uint32 {
	return (*c).LastConfirmedBlock.BalanceOf((*c).Address)
}

// Any gold received in the last confirmed block or before
func (c *Client) AvailableGold() uint32 {
	var pendingSpent uint32 = 0
	for _, tx := range (*c).PendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return c.ConfirmedBalance() - pendingSpent
}

// Broadcasts a transaction from the client giving gold to the clients
func (c *Client) PostTransaction(outputs []Output, fee uint32) *Transaction {

	(*c).mu.Lock()
	defer (*c).mu.Unlock()

	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > c.AvailableGold() {
		// modify here
		panic(`Account doesn't have enough balance for transaction`)
	}
	// add data to the constructor
	tx, _ := NewTransaction((*c).Address, (*c).Nonce, (*c).PubKey, nil, fee, outputs, nil)

	tx.Sign((*c).PrivKey)
	(*c).PendingOutgoingTransactions[tx.Id()] = tx
	(*c).Nonce++
	data, _ := TransactionToBytes(tx)
	(*c).Net.Broadcast(POST_TRANSACTION, data)

	return tx
}

// Validates and adds a block to the list of blocks, possibly
// updating the head of the blockchain.
func (c *Client) ReceiveBlock(b Block) *Block {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()

	block := &b
	blockId, _ := block.GetHash()

	if _, received := (*c).Blocks[blockId]; received {
		return nil
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		c.Log(fmt.Sprintf("Block %v does not have a valid proof\n", blockId))
		return nil
	}

	//var prevBlock *Block = nil
	prevBlock, received := (*c).Blocks[(*block).PrevBlockHash]
	if !received && !block.IsGenesisBlock() {

		stuckBlocks, received := (*c).PendingBlocks[(*block).PrevBlockHash]
		if !received {
			c.RequestMissingBlock(block)
			// TODO: Define a set
			stuckBlocks = NewSet[*Block]()
		}
		stuckBlocks.Add(block)
		(*c).PendingBlocks[(*block).PrevBlockHash] = stuckBlocks
		return nil

	}

	if !block.IsGenesisBlock() {
		if !block.Rerun(prevBlock) {
			return nil
		}
	}

	blockId, _ = block.GetHash()
	(*c).Blocks[blockId] = block

	if (*(*c).LastBlock).ChainLength < (*block).ChainLength {
		(*c).LastBlock = block
		c.SetLastConfirmed()
	}

	unstuckBlocks, received := (*c).PendingBlocks[blockId]
	var unstuckBlocksArr []*Block
	if received {
		unstuckBlocksArr = unstuckBlocks.ToArray()
	}

	delete((*c).PendingBlocks, blockId)

	for _, uBlock := range unstuckBlocksArr {
		c.Log(fmt.Sprintf("processing unstuck block %v", uBlock.GetHashStr()))
		// Need to change the "" into empty []byte
		go c.ReceiveBlock(*uBlock)
	}
	c.Log(fmt.Sprintf("block %s received", block.GetHashStr()))
	return block
}

func (c *Client) ReceiveBlockBytes(bs []byte) *Block {

	block, err := BytesToBlock(bs)
	if err != nil {
		panic("Failed to deseralize block")
	}

	return c.ReceiveBlock(*block)
}

// Request the previous block from the network.
// convert []byte into string
func (c *Client) RequestMissingBlock(block *Block) {
	c.Log(fmt.Sprintf("Asking for missing block: %v", (*block).PrevBlockHash))
	var msg = Message{(*c).Address, (*block).PrevBlockHash}
	jsonByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("RequestMissingBlock() Marshal Panic:")
		panic(err)
	}
	(*c).Net.Broadcast(MISSING_BLOCK, jsonByte)
}

// Resend any transactions in the pending list
func (c *Client) ResendPendingTransactions() {
	for _, tx := range (*c).PendingOutgoingTransactions {
		jsonByte, err := json.Marshal(*tx)
		if err != nil {
			fmt.Println("ResendPendingTransactions() Marshal Panic:")
			panic(err)
		}
		(*c).Net.Broadcast(POST_TRANSACTION, jsonByte)
	}
}

// Takes an object representing a request for a missing block
func (c *Client) ProvideMissingBlock(data []byte) {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("ProvideMissingBlock() unmarshal Panic:")
		panic(err)
	}
	if val, received := (*c).Blocks[msg.PrevBlockHash]; received {
		c.Log(fmt.Sprintf("Providing missing block %v", val.GetHashStr()))
		data, err := BlockToBytes(val)
		if err != nil {
			fmt.Println("ProvideMissingBlock() Marshal Panic:")
			panic(err)
		}
		(*c).Net.SendMessage(msg.Address, PROOF_FOUND, data)
	}
}

// Sets the last confirmed block according to the most accepted block and also
// updating pending transactions according to this block.
func (c *Client) SetLastConfirmed() {
	block := (*c).LastBlock
	confirmedBlockHeight := uint32(0)
	if (*block).ChainLength > CONFIRMED_DEPTH {
		confirmedBlockHeight = (*block).ChainLength - CONFIRMED_DEPTH
	}
	for (*block).ChainLength > confirmedBlockHeight {
		block = (*c).Blocks[(*block).PrevBlockHash]
	}
	(*c).LastConfirmedBlock = block
	for id, tx := range (*c).PendingOutgoingTransactions {
		if (*c).LastConfirmedBlock.Contains(tx) {
			delete((*c).PendingOutgoingTransactions, id)
		}
	}
}

// Utility method that displays all confirmed balances for all clients
func (c *Client) ShowAllBalances() {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()
	fmt.Printf("Showing balances:")
	for id, balance := range (*(*c).LastConfirmedBlock).Balances {
		fmt.Printf("	%v", id)
		fmt.Printf("	%v", balance)
		fmt.Println("")
	}
}

// Logs messages to stdout
func (c *Client) Log(msg string) {
	name := (*c).Address[0:10]
	if len((*c).Name) > 0 {
		name = (*c).Name
	}
	fmt.Printf("	%s", name)
	fmt.Printf("	%s\n", msg)
}

// Print out the blocks in the blockchain from the current head to the genesis block.
func (c *Client) ShowBlockchain() {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()
	block := (*c).LastBlock
	fmt.Println("BLOCKCHAIN:")
	for block != nil {
		blockId, _ := block.GetHash()
		fmt.Println(blockId)
		block = (*c).Blocks[(*block).PrevBlockHash]
	}
}

func NewClient(name string, Net *FakeNet, startingBlock *Block, keyPair *rsa.PrivateKey) *Client {
	var c Client
	c.Net = Net
	c.Name = name

	if keyPair == nil {
		c.PrivKey, c.PubKey, _ = GenerateKeypair()
	} else {
		c.PrivKey = keyPair
		c.PubKey = &keyPair.PublicKey
	}
	c.Address = GenerateAddress(c.PubKey)
	c.Nonce = 0

	c.PendingOutgoingTransactions = make(map[string]*Transaction)
	c.PendingReceivedTransactions = make(map[string]*Transaction)
	c.Blocks = make(map[string]*Block)
	c.PendingBlocks = make(map[string]*Set[*Block])

	if startingBlock != nil {
		c.SetGenesisBlock(startingBlock)
	}

	c.Emitter = emission.NewEmitter()
	c.Emitter.On(PROOF_FOUND, c.ReceiveBlockBytes)
	c.Emitter.On(MISSING_BLOCK, c.ProvideMissingBlock)
	return &c
}

func (c *Client) GetAddress() string {
	return (*c).Address
}
func (c *Client) GetEmitter() *emission.Emitter {
	return (*c).Emitter
}
