package blockchain

import (
	"crypto/rsa"
	"fmt"
)

// Miners are clients, but they also mine blocks looking for "proofs".
type Miner struct {
	Name                        string
	Address                     string
	Nonce                       uint32
	Net                         *FakeNet
	KeyPair                     *rsa.PrivateKey
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	Blocks                      map[string]*Block
	PendingBlocks               map[string]*Block
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceiveBlock                *Block
	CurrentBlock                *Block
	MiningRounds                uint32
	BlockChain                  *BlockChain
	CBTXS                       *Set[*Transaction]
	NBTXS                       *Set[*Transaction]
	TXSET                       *Set[*Transaction]
	TX                          *Set[*Transaction]
}

func newMiner(name string, Net *FakeNet, startingBlock *Block, keyPair *rsa.PrivateKey) *Miner {
	var m *Miner
	m.Net = Net
	m.Name = name

	if keyPair == nil {
		m.KeyPair = utils.GenerateKeypair()
	} else {
		m.KeyPair = keyPair
	}
	m.Address = utils.GenerateAddress(&m.KeyPair.PublicKey)
	m.Nonce = 0

	m.PendingOutgoingTransactions = make(map[string]*Transaction)
	m.PendingReceivedTransactions = make(map[string]*Transaction)
	m.Blocks = make(map[string]*Block)
	m.PendingBlocks = make(map[string]*Block)

	if startingBlock != nil {
		m.setGenesisBlock(startingBlock)
	}
	m.MiningRounds = NUM_ROUNDS_MINING

	return &m
}

func (m *Miner) setGenesisBlock(startingBlock *Block) {
	if m.LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain")
	}
	m.LastConfirmedBlock = startingBlock
	m.LastBlock = startingBlock
	blockId, _ := startingBlock.GetHash()
	m.Blocks[blockId] = startingBlock
}

// Starts listeners and begins mining
func (m *Miner) initialize() {
	m.startNewSearch()
	m.on(START_MINING, m.findProof)
	m.on(POST_TRANSACTION, m.addTransaction)
}

func (m *Miner) startNewSearch() {
	//TODO assign currentBlock
	m.currentBlock = Blockchain.makeBlock(m.address, m.lastBlock)

	for _, transaction := range m.TXSET {
		m.addTransaction(transaction)
	}
	m.CurrentBlock.Proof = 0
}

// Broadcast the block, with a valid proof included
func (m *Miner) announceProof() {
	m.Net.broadcast(PROOF_FOUND, m.CurrentBlock)
}

// Looks for a "proof".
func (m *Miner) findProof(oneAndDone bool) {
	if oneAndDone == nil {
		oneAndDone = false
	}
	pausePoint := m.CurrentBlock.Proof + m.MiningRounds

	for m.CurrentBlock.Proof < pausePoint {
		if m.CurrentBlock.hasValidProof() == true {
			fmt.Printf("found proof for block %v", m.CurrentBlock.ChainLength)
			fmt.Printf(": %v\n", m.CurrentBlock.Proof)
			m.announceProof()
			m.minerreceiveBlock(&m.CurrentBlock)
			m.startNewSearch()
			break
		}
	}
	m.CurrentBlock.Proof++
	// If we are testing, don't continue the search
	// TODO
	if oneAndDone == false {
		//setTimeout(() => m.emit(Blockchain.START_MINING), 0)
	}
}

// This function should determine what transactions need to be added or deleted.
func (m *Miner) syncTransaction(newBlock *Block) *Set {
	var cb = m.CurrentBlock
	m.CBTXS = &NewSet()
	m.NBTXS = &NewSet()

	for newBlock.ChainLength > cb.ChainLength {
		for _, transaction := range newBlock.Transactions {
			m.NBTXS.Add(transaction)
			newBlock = m.Blocks[newBlock.PrevBlockHash]
		}
	}
	currentBlockId, _ := cb.GetHash()
	newBlockId, _ := newBlock.GetHash()
	for cb != nil && currentBlockId != newBlockId {
		for _, transaction := range cb.Transactions {
			m.CBTXS.Add(transaction)
		}
		for _, transaction := range newBlock.Transactions {
			m.NBTXS.Add(transaction)
		}
		newBlock = m.Blocks[newBlock.PrevBlockHash]
		cb = m.Blocks[cb.PrevBlockHash]
	}

	for _, transaction := range m.NBTXS {
		m.CBTXS.Remove(transaction)
	}

	return m.CBTXS
}

// Returns false if transaction is not accepted.Otherwise add the
// transaction to the current block.
func (m *Miner) addTransaction(tx *Transaction) bool {
	//TODO
	tx = m.BlockChain.makeTransaction(tx)
	return m.CurrentBlock.addTransaction(tx, m)
}

func (m *Miner) inheritedpostTransaction(outputs []Output, fee int) *Transaction {

	if fee < 0 {
		fee = m.BlockChain.DefaultTxFee
	}
	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > m.availableGold() {
		panic(`Account doesn't have enough balance for transaction`)
	}

	tx := NewTransaction(m.Address, m.Nonce, &m.KeyPair.PublicKey, nil, fee, outputs, nil)

	tx.Sign(m.KeyPair)

	m.PendingOutgoingTransactions[tx.Id()] = tx

	m.Nonce++

	return tx
}

func (m *Miner) minerpostTransaction(args ...interface{}) bool {

	var totalArgs []Output

	for _, arg := range args {
		s = append(totalArgs, arg)

	}

	m.TX = inheritedpostTransaction(totalArgs, -1)
	return m.addTransaction(m.TX)
}

func (m *Miner) minerreceiveBlock(s *Block) {

	block := s
	if s == nil {
		block = BytesToBlock(bs)
	}
	blockId, _ = block.GetHash()
	if _, received := m.Blocks[blockId]; received {
		return nil
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		fmt.Printf("Block %v does not have a valid proof", blockId)
		return nil
	}

	var prevBlock *Block = nil

	prevBlock, received := m.Blocks[block.PrevBlockHash]
	if !received {
		if !prevBlock.IsGenesisBlock() {
			stuckBlocks, received := m.PendingBlocks[block.PrevBlockHash]
			if !received {
				m.requestMissingBlock(*block)
				stuckBlocks = make([]*Block, 10)
			}
			stuckBlocks = append(stuckBlocks, block)
			m.PendingBlocks[block.PrevBlockHash] = stuckBlocks
			return nil
		}
	}

	if !block.IsGenesisBlock() {
		if !block.rerun(prevBlock) {
			return nil
		}
	}

	blockId, _ = block.GetHash()
	m.Blocks[blockId] = block

	if m.LastBlock.ChainLength < block.ChainLength {
		m.LastBlock = block
		m.setLastConfirmed()
	}

	unstuckBlocks := make([]*Block, 0)
	if val, received := m.PendingBlocks[blockId]; received {
		unstuckBlocks = val
	}

	delete(m.PendingBlocks, blockId)

	for _, uBlock := range unstuckBlocks {
		fmt.Printf("processing unstuck block %v", blockId)
		//TODO
		m.receiveBlock(uBlock, "")
	}

	var b *Block = block

	if b == nil {
		return nil
	} else {
		if m.CurrentBlock == true && b.ChainLength >= m.CurrentBlock.ChainLength {
			fmt.Printf("cutting over to the new chain \n")
			m.TXSET = m.syncTransactions(b)
			m.startNewSearch(m.TXSET)
		}

	}

}
