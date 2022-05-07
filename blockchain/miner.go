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
	PrivKey                     *rsa.PrivateKey
	PubKey                      *rsa.PublicKey
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	Blocks                      map[string]*Block
	PendingBlocks               map[string]*Block
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceiveBlock                *Block
	CurrentBlock                *Block
	MiningRounds                uint32
	Config                      BlockchainConfig
	CBTXS                       *Set[*Transaction]
	NBTXS                       *Set[*Transaction]
	TXSET                       *Set[*Transaction]
	TX                          *Set[*Transaction]
}

func NewMiner(name string, Net *FakeNet, startingBlock *Block, keyPair *rsa.PrivateKey) *Miner {
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

	m.PendingOutgoingTransactions = make(map[string]*Transaction)
	m.PendingReceivedTransactions = make(map[string]*Transaction)
	m.Blocks = make(map[string]*Block)
	m.PendingBlocks = make(map[string]*Block)

	if startingBlock != nil {
		m.SetGenesisBlock(startingBlock)
	}
	m.MiningRounds = NUM_ROUNDS_MINING

	return &m
}

func (m *Miner) SetGenesisBlock(startingBlock *Block) {
	if m.LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain")
	}
	m.LastConfirmedBlock = startingBlock
	m.LastBlock = startingBlock
	blockId, _ := startingBlock.GetHash()
	m.Blocks[blockId] = startingBlock
}

// Starts listeners and begins mining
func (m *Miner) Initialize() {
	//TODO
	/*
		m.startNewSearch()
		m.on(START_MINING, m.findProof)
		m.on(POST_TRANSACTION, m.addTransaction)
	*/
}

func (m *Miner) StartNewSearch() {
	//TODO assign currentBlock
	target := CalculateTarget(POW_LEADING_ZEROES)
	m.CurrentBlock = NewBlock((*m).Address, (*m).LastBlock, target, (*m).Config.coinbaseAmount)
	//Blockchain.makeBlock(m.address, m.lastBlock)

	txList := (*m).TXSET.ToArray()
	for _, transaction := range txList {
		(*m).AddTransaction(transaction)
	}
	m.CurrentBlock.Proof = 0
}

// Broadcast the block, with a valid proof included
func (m *Miner) AnnounceProof() {
	(*m).Net.Broadcast(PROOF_FOUND, m.CurrentBlock)
}

// Looks for a "proof".
func (m *Miner) FindProof(oneAndDone bool) {

	pausePoint := (*m).CurrentBlock.Proof + (*m).MiningRounds

	for (*m).CurrentBlock.Proof < pausePoint {
		if (*m).CurrentBlock.hasValidProof() == true {
			fmt.Printf("found proof for block %v", m.CurrentBlock.ChainLength)
			fmt.Printf(": %v\n", m.CurrentBlock.Proof)
			(*m).AnnounceProof()
			(*m).MinerreceiveBlock((*m).CurrentBlock)
			(*m).StartNewSearch()
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
func (m *Miner) SyncTransaction(newBlock *Block) *Set[*Transaction] {
	var cb = m.CurrentBlock
	m.CBTXS = NewSet[*Transaction]()
	m.NBTXS = NewSet[*Transaction]()

	for newBlock.ChainLength > cb.ChainLength {
		for _, transaction := range newBlock.Transactions {
			m.NBTXS.Add(&transaction)
			newBlock = m.Blocks[newBlock.PrevBlockHash]
		}
	}
	currentBlockId, _ := cb.GetHash()
	newBlockId, _ := newBlock.GetHash()
	for cb != nil && currentBlockId != newBlockId {
		for _, transaction := range cb.Transactions {
			m.CBTXS.Add(&transaction)
		}
		for _, transaction := range newBlock.Transactions {
			m.NBTXS.Add(&transaction)
		}
		newBlock = m.Blocks[newBlock.PrevBlockHash]
		cb = m.Blocks[cb.PrevBlockHash]
	}

	NBTXS_List := (*m).NBTXS.ToArray()
	for _, transaction := range NBTXS_List {
		(*m).CBTXS.Remove(transaction)
	}

	return (*m).CBTXS
}

// Returns false if transaction is not accepted.Otherwise add the
// transaction to the current block.
func (m *Miner) AddTransaction(tx *Transaction) bool {
	//TODO
	//tx = m.BlockChain.makeTransaction(tx)
	return (*m).CurrentBlock.AddTransaction(tx)
}

func (m *Miner) InheritedpostTransaction(outputs []Output, fee uint32) *Transaction {

	if fee < 0 {
		fee = (*m).Config.defaultTxFee
	}
	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	if total > m.CurrentBlock.BalanceOf((*m).Address) {
		panic(`Account doesn't have enough balance for transaction`)
	}

	tx, _ := NewTransaction((*m).Address, (*m).Nonce, (*m).PubKey, nil, fee, outputs, nil)
	//NewTransaction(m.Address, m.Nonce, &m.KeyPair.PublicKey, nil, fee, outputs, nil)

	tx.Sign((*m).PrivKey)

	m.PendingOutgoingTransactions[tx.Id()] = tx

	m.Nonce++

	return tx
}

/* TODO
func (m *Miner) MinerpostTransaction(args ...interface{}) bool {

	var totalArgs []Output

	for _, arg := range args {
		totalArgs = append(totalArgs, arg)
	}

	m.TX = m.InheritedpostTransaction(totalArgs, -1)
	return m.AddTransaction(m.TX)
}*/

func (m *Miner) MinerreceiveBlock(s *Block) {

	block := s

	blockId, _ := block.GetHash()
	if _, received := m.Blocks[blockId]; received {
		return
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		fmt.Printf("Block %v does not have a valid proof", blockId)
		return
	}

	var prevBlock *Block = nil

	prevBlock, received := m.Blocks[block.PrevBlockHash]
	if !received {
		if !prevBlock.IsGenesisBlock() {
			/* TODO
			stuckBlocks, received := m.PendingBlocks[block.PrevBlockHash]
			if !received {
				m.RequestMissingBlock(*block)
				stuckBlocks = make([]*Block, 10)
			}
			stuckBlocks = append(stuckBlocks, block)
			m.PendingBlocks[block.PrevBlockHash] = stuckBlocks
			*/
			return
		}
	}

	if !block.IsGenesisBlock() {
		if !block.Rerun(prevBlock) {
			return
		}
	}

	blockId, _ = block.GetHash()
	m.Blocks[blockId] = block

	if m.LastBlock.ChainLength < block.ChainLength {
		m.LastBlock = block
		// TODO
		//	m.SetLastConfirmed()
	}

	unstuckBlocks := NewSet[*Block]()
	if val, received := m.PendingBlocks[blockId]; received {
		unstuckBlocks.Add(val)
	}

	delete(m.PendingBlocks, blockId)

	//unstuckBlocksArr := unstuckBlocks.ToArray()
	//for _, uBlock := range unstuckBlocksArr {
	//	fmt.Printf("processing unstuck block %v", blockId)
	//TODO
	//m.ReceiveBlock(uBlock, "")
	//}

	var b *Block = block

	if b == nil {
		return
	} else {
		if m.CurrentBlock != nil && b.ChainLength >= m.CurrentBlock.ChainLength {
			fmt.Printf("cutting over to the new chain \n")
			// TODO
			//m.TXSET = m.SyncTransactions(b)
			m.StartNewSearch()
		}

	}

}
