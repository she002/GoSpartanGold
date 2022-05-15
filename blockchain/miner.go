package blockchain

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/chuckpreslar/emission"
)

// Miners are clients, but they also mine blocks looking for "proofs".
type Miner struct {
	// The variables that are same as Client
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

	// Miner's specific variables
	CurrentBlock *Block
	MiningRounds uint32
	Transactions *Set[*Transaction]
}

func NewMiner(name string, Net *FakeNet, miningRounds uint32, startingBlock *Block, keyPair *rsa.PrivateKey, config BlockchainConfig) *Miner {
	var m Miner
	m.Net = Net
	m.Name = name

	if keyPair == nil {
		m.PrivKey, m.PubKey, _ = GenerateKeypair()
	} else {
		m.PrivKey = keyPair
		m.PubKey = &keyPair.PublicKey
	}

	m.Address = GenerateAddress(m.PubKey)
	m.Nonce = 0
	m.Config = config
	m.PendingOutgoingTransactions = make(map[string]*Transaction)
	m.PendingReceivedTransactions = make(map[string]*Transaction)
	m.Blocks = make(map[string]*Block)
	m.PendingBlocks = make(map[string]*Set[*Block])

	if startingBlock != nil {
		m.SetGenesisBlock(startingBlock)
	}

	m.Emitter = emission.NewEmitter()
	m.Emitter.On(PROOF_FOUND, m.ReceiveBlockBytes)
	m.Emitter.On(MISSING_BLOCK, m.ProvideMissingBlock)

	m.MiningRounds = miningRounds

	m.Transactions = NewSet[*Transaction]()

	return &m
}

func (m *Miner) SetGenesisBlock(startingBlock *Block) {
	if (*m).LastBlock != nil {
		panic("Cannot set starting block for existing blockchain")
	}
	(*m).LastConfirmedBlock = startingBlock
	(*m).LastBlock = startingBlock
	blockId, _ := startingBlock.GetHash()
	(*m).Blocks[blockId] = startingBlock
}

// Starts listeners and begins mining
func (m *Miner) Initialize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartNewSearch(nil)

	(*m).Emitter.On(START_MINING, m.FindProof)
	(*m).Emitter.On(POST_TRANSACTION, m.AddTransactionBytes)

	go (*m).Emitter.Emit(START_MINING, false)
}

func (m *Miner) StartNewSearch(txSet *Set[*Transaction]) {

	//TODO assign currentBlock
	target := CalculateTarget(POW_LEADING_ZEROES)
	(*m).CurrentBlock = NewBlock((*m).Address, (*m).LastBlock, target, (*m).Config.coinbaseAmount)
	//Blockchain.makeBlock(m.address, m.lastBlock)

	if txSet == nil {
		txSet = NewSet[*Transaction]()
	}

	txList := txSet.ToArray()

	for _, transaction := range txList {
		(*m).Transactions.Add(transaction)
	}

	transactionsArr := (*m).Transactions.ToArray()
	for _, transaction := range transactionsArr {
		(*m).CurrentBlock.AddTransaction(transaction)
	}
	(*m).Transactions.Clear()

	(*m).CurrentBlock.Proof = 0

}

// Looks for a "proof".
func (m *Miner) FindProof(oneAndDone bool) {

	(*m).mu.Lock()
	defer (*m).mu.Unlock()

	pausePoint := (*m).CurrentBlock.Proof + (*m).MiningRounds

	for (*m).CurrentBlock.Proof < pausePoint {
		if (*m).CurrentBlock.hasValidProof() {
			m.Log(fmt.Sprintf("found proof for block %d: %d", (*m).CurrentBlock.ChainLength, (*m).CurrentBlock.Proof))
			m.AnnounceProof()
			go m.ReceiveBlock(*(*m).CurrentBlock)
			break
		}
		(*m).CurrentBlock.Proof++
	}

	// If we are testing, don't continue the search
	// TODO
	if !oneAndDone {
		go (*m).Emitter.Emit(START_MINING, false)
	}
}

// Broadcast the block, with a valid proof included
func (m *Miner) AnnounceProof() {

	data, err := BlockToBytes((*m).CurrentBlock)
	if err != nil {
		fmt.Println("AnnounceProof() Marshal Panic:")
		panic(err)
	}
	(*m).Net.Broadcast(PROOF_FOUND, data)
}

// Validates and adds a block to the list of blocks, possibly
// updating the head of the blockchain.
func (m *Miner) ReceiveBlock(b Block) *Block {
	(*m).mu.Lock()
	defer (*m).mu.Unlock()

	block := &b
	blockId, _ := block.GetHash()

	if _, received := (*m).Blocks[blockId]; received {
		return nil
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		m.Log(fmt.Sprintf("Block %v does not have a valid proof\n", blockId))
		return nil
	}

	//var prevBlock *Block = nil
	prevBlock, received := (*m).Blocks[(*block).PrevBlockHash]
	if !received && !block.IsGenesisBlock() {

		stuckBlocks, received := (*m).PendingBlocks[(*block).PrevBlockHash]
		if !received {
			m.RequestMissingBlock(block)
			// TODO: Define a set
			stuckBlocks = NewSet[*Block]()
		}
		stuckBlocks.Add(block)
		(*m).PendingBlocks[block.PrevBlockHash] = stuckBlocks
		return nil

	}

	if !block.IsGenesisBlock() {
		if !block.Rerun(prevBlock) {
			return nil
		}
	}

	blockId, _ = block.GetHash()
	(*m).Blocks[blockId] = block

	if (*(*m).LastBlock).ChainLength < (*block).ChainLength {
		(*m).LastBlock = block
		m.SetLastConfirmed()
	}

	unstuckBlocks, received := (*m).PendingBlocks[blockId]
	var unstuckBlocksArr []*Block
	if received {
		unstuckBlocksArr = unstuckBlocks.ToArray()
	}

	delete((*m).PendingBlocks, blockId)

	for _, uBlock := range unstuckBlocksArr {
		m.Log(fmt.Sprintf("processing unstuck block %v", uBlock.GetHashStr()))
		// Need to change the "" into empty []byte
		go m.ReceiveBlock(*uBlock)
	}
	m.Log(fmt.Sprintf("block %s received", block.GetHashStr()))

	if (*m).CurrentBlock != nil && (*block).ChainLength >= (*m).CurrentBlock.ChainLength {
		m.Log("Cutting over to new chain")
		txSet := m.SyncTransaction(block)
		m.StartNewSearch(txSet)
	}

	return block
}

func (m *Miner) ReceiveBlockBytes(bs []byte) *Block {

	block, err := BytesToBlock(bs)
	if err != nil {
		panic("Failed to deseralize block")
	}

	return m.ReceiveBlock(*block)
}

// This function should determine what transactions need to be added or deleted.
func (m *Miner) SyncTransaction(newBlock *Block) *Set[*Transaction] {

	cb := (*m).CurrentBlock
	cbTxs := NewSet[*Transaction]()
	nbTxs := NewSet[*Transaction]()

	for newBlock.ChainLength > cb.ChainLength {
		for _, transaction := range newBlock.Transactions {
			nbTxs.Add(&transaction.Tx)
		}
		newBlock = (*m).Blocks[newBlock.PrevBlockHash]
	}

	currentBlockId, _ := cb.GetHash()
	newBlockId, _ := newBlock.GetHash()
	for currentBlockId != newBlockId {
		for _, transaction := range cb.Transactions {
			cbTxs.Add(&transaction.Tx)
		}
		for _, transaction := range newBlock.Transactions {
			nbTxs.Add(&transaction.Tx)
		}
		newBlock = (*m).Blocks[newBlock.PrevBlockHash]
		cb = (*m).Blocks[cb.PrevBlockHash]

		if cb != nil {
			currentBlockId, _ = cb.GetHash()
			newBlockId, _ = newBlock.GetHash()
		} else {
			break
		}
	}

	nbTxsArr := nbTxs.ToArray()
	for _, transaction := range nbTxsArr {
		cbTxs.Remove(transaction)
	}

	return cbTxs
}

// Returns false if transaction is not accepted.Otherwise add the
// transaction to the current block.
func (m *Miner) AddTransaction(tx *Transaction) {
	(*m).mu.Lock()
	defer (*m).mu.Unlock()
	(*m).Transactions.Add(tx)
}

func (m *Miner) AddTransactionBytes(data []byte) {

	tx, err := BytesToTransaction(data)
	if err != nil {
		panic("Failed to deserialize data")
	}
	m.AddTransaction(tx)
}

// The amount of gold available to the client looking at the last confirmed block
func (m *Miner) ConfirmedBalance() uint32 {
	return (*m).LastConfirmedBlock.BalanceOf((*m).Address)
}

// Any gold received in the last confirmed block or before
func (m *Miner) AvailableGold() uint32 {
	var pendingSpent uint32 = 0
	for _, tx := range (*m).PendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return m.ConfirmedBalance() - pendingSpent
}

func (m *Miner) PostTransaction(outputs []Output, fee uint32) {

	(*m).mu.Lock()

	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > m.AvailableGold() {
		// modify here
		panic(`Account doesn't have enough balance for transaction`)
	}
	// add data to the constructor
	tx, _ := NewTransaction((*m).Address, (*m).Nonce, (*m).PubKey, nil, fee, outputs, nil)

	tx.Sign((*m).PrivKey)
	(*m).PendingOutgoingTransactions[tx.Id()] = tx
	(*m).Nonce++
	data, _ := TransactionToBytes(tx)
	(*m).Net.Broadcast(POST_TRANSACTION, data)
	(*m).mu.Unlock()

	m.AddTransaction(tx)
}

// Request the previous block from the network.
// convert []byte into string
func (m *Miner) RequestMissingBlock(block *Block) {
	m.Log(fmt.Sprintf("Asking for missing block: %v", block.PrevBlockHash))
	var msg = Message{(*m).Address, (*block).PrevBlockHash}
	jsonByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("RequestMissingBlock() Marshal Panic:")
		panic(err)
	}
	(*m).Net.Broadcast(MISSING_BLOCK, jsonByte)
}

// Takes an object representing a request for a missing block
func (m *Miner) ProvideMissingBlock(data []byte) {
	(*m).mu.Lock()
	defer (*m).mu.Unlock()

	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("ProvideMissingBlock() unmarshal Panic:")
		panic(err)
	}
	if val, received := (*m).Blocks[msg.PrevBlockHash]; received {
		m.Log(fmt.Sprintf("Providing missing block %v", val.GetHashStr()))
		data, err := BlockToBytes(val)
		if err != nil {
			fmt.Println("ProvideMissingBlock() Marshal Panic:")
			panic(err)
		}
		(*m).Net.SendMessage(msg.Address, PROOF_FOUND, data)
	}
}

// Sets the last confirmed block according to the most accepted block and also
// updating pending transactions according to this block.
func (m *Miner) SetLastConfirmed() {
	block := (*m).LastBlock
	confirmedBlockHeight := uint32(0)
	if (*block).ChainLength > CONFIRMED_DEPTH {
		confirmedBlockHeight = (*block).ChainLength - CONFIRMED_DEPTH
	}
	for (*block).ChainLength > confirmedBlockHeight {
		block = (*m).Blocks[block.PrevBlockHash]
	}
	(*m).LastConfirmedBlock = block
	for id, tx := range (*m).PendingOutgoingTransactions {
		if (*m).LastConfirmedBlock.Contains(tx) {
			delete((*m).PendingOutgoingTransactions, id)
		}
	}
}

// Utility method that displays all confirmed balances for all clients
func (m *Miner) ShowAllBalances() {

	fmt.Printf("Showing balances:")
	for id, balance := range (*m).LastConfirmedBlock.Balances {
		fmt.Printf("	%v", id)
		fmt.Printf("	%v", balance)
		fmt.Println("")
	}
}

// Print out the blocks in the blockchain from the current head to the genesis block.
func (m *Miner) ShowBlockchain() {

	block := (*m).LastBlock
	fmt.Println("BLOCKCHAIN:")
	for block != nil {
		blockId, _ := block.GetHash()
		fmt.Println(blockId)
		block = (*m).Blocks[(*block).PrevBlockHash]
	}
}

// Logs messages to stdout
func (m *Miner) Log(msg string) {
	name := (*m).Address[0:10]
	if len((*m).Name) > 0 {
		name = (*m).Name
	}
	fmt.Printf("	%s", name)
	fmt.Printf("	%s\n", msg)
}

func (m *Miner) GetAddress() string {
	return (*m).Address
}
func (m *Miner) GetEmitter() *emission.Emitter {
	return (*m).Emitter
}
