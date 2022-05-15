package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type TransactionType struct {
	Id string
	Tx Transaction
}

type NextNonceType struct {
	Id    string
	Nonce uint32
}

type BalanceType struct {
	Id      string
	Balance uint32
}

type Block struct {
	PrevBlockHash  string
	Target         big.Int
	Proof          uint32
	Balances       []BalanceType
	NextNonce      []NextNonceType
	Transactions   []TransactionType
	ChainLength    uint32
	Timestamp      time.Time
	RewardAddr     string
	CoinbaseReward uint32
}

func (block *Block) FindTransactionIndex(id string) int {
	index := int(-1)
	for i, v := range block.Transactions {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func (block *Block) FindNextNonceIndex(id string) int {
	index := int(-1)
	for i, v := range block.NextNonce {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func (block *Block) FindBalanceIndex(id string) int {
	index := int(-1)
	for i, v := range block.Balances {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func NewBlock(rewardAddr string, prevBlock *Block, target *big.Int, coinbaseReward uint32) *Block {
	var block Block
	block.Target = *target
	block.Proof = 0

	if prevBlock != nil {
		hashHexStr, err := prevBlock.GetHash()
		if err != nil {
			fmt.Println(err.Error())
			return nil
		} else {
			block.PrevBlockHash = hashHexStr
		}
	}

	// copy Balances from previous block
	block.Balances = make([]BalanceType, 0)
	if prevBlock != nil && (*prevBlock).Balances != nil {
		block.Balances = append(block.Balances, (*prevBlock).Balances...)
	}

	// copy NextNounce from previous block
	block.NextNonce = make([]NextNonceType, 0)
	if prevBlock != nil && (*prevBlock).NextNonce != nil {
		block.NextNonce = append(block.NextNonce, (*prevBlock).NextNonce...)
	}

	block.Transactions = make([]TransactionType, 0)

	block.ChainLength = 0
	if prevBlock != nil {
		block.ChainLength = (*prevBlock).ChainLength + 1
	}

	block.Timestamp = time.Now()
	block.RewardAddr = rewardAddr
	block.CoinbaseReward = coinbaseReward
	return &block
}

func (block *Block) ToString() string {
	var blockStr string

	blockStr = fmt.Sprintf("PrevBlockHash: %s\n", (*block).PrevBlockHash)
	blockStr = blockStr + fmt.Sprintf("Target: %x\n", &(*block).Target)
	blockStr = blockStr + fmt.Sprintf("Proof: %d\n", (*block).Proof)
	blockStr = blockStr + fmt.Sprintf("ChainLength: %d\n", (*block).ChainLength)
	blockStr = blockStr + fmt.Sprintf("Timestamp: %s\n", (*block).Timestamp.GoString())
	blockStr = blockStr + fmt.Sprintf("RewardAddr: %s\n", (*block).RewardAddr)
	blockStr = blockStr + fmt.Sprintf("CoinbaseReward: %d\n", (*block).CoinbaseReward)

	balanceStr := "Balancecs: [\n"
	for _, v := range (*block).Balances {
		balanceStr = balanceStr + fmt.Sprintf("\t%s: %d\n", v.Id, v.Balance)
	}
	balanceStr = balanceStr + "]\n"

	nextNonceStr := "nextNonceStr: [\n"
	for _, v := range (*block).NextNonce {
		nextNonceStr = nextNonceStr + fmt.Sprintf("\t%s: %d\n", v.Id, v.Nonce)
	}
	nextNonceStr = nextNonceStr + "]\n"

	transactionStr := "Transactions: [\n"
	for _, v := range (*block).Transactions {
		transactionStr = transactionStr + fmt.Sprintf("\t%s\n", v.Id)
	}
	transactionStr = transactionStr + "]\n"

	blockStr = blockStr + balanceStr + nextNonceStr + transactionStr

	return blockStr
}

func BlockToBytes(block *Block) ([]byte, error) {
	data, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func BytesToBlock(data []byte) (*Block, error) {
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (block *Block) GetHash() (string, error) {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil

	blockData, err := BlockToBytes(&block4hash)
	var blockHash [32]byte
	if err == nil {
		blockHash = sha256.Sum256(blockData)
		return hex.EncodeToString(blockHash[:]), nil
	} else {
		return "", err
	}
}

func (block *Block) GetHashStr() string {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil

	blockData, err := BlockToBytes(&block4hash)
	var blockHash [32]byte
	if err == nil {
		blockHash = sha256.Sum256(blockData)
		return hex.EncodeToString(blockHash[:])
	} else {
		return ""
	}
}

func (block *Block) IsGenesisBlock() bool {
	return block.ChainLength == 0
}

func (block *Block) hasValidProof() bool {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil
	data, err := BlockToBytes(&block4hash)
	if err != nil {
		fmt.Printf("Failed to convert Block to Byte array")
		return false
	}
	block_hash := sha256.Sum256(data)
	block_value := big.NewInt(0)
	block_value.SetBytes(block_hash[:])

	return block_value.Cmp(&(*block).Target) < 0
}

func (block *Block) AddTransaction(tx *Transaction) bool {
	if (*block).Contains(tx) {
		fmt.Printf("Duplicate transaction %s", tx.Id())
		return false
	} else if (*tx).Sig == nil {
		fmt.Printf("Unsigned transaction %s", tx.Id())
		return false
	} else if !tx.VerifySignature() {
		fmt.Printf("Invalid signature for transaction %s", tx.Id())
		return false
	} else if !block.SufficientFund(tx) {
		fmt.Printf("Insufficient fund for transaction %s", tx.Id())
		return false
	}

	nextNonceIndex := block.FindNextNonceIndex((*tx).Info.From)
	if nextNonceIndex == -1 {
		addNextNonce := NextNonceType{Id: (*tx).Info.From, Nonce: 0}
		(*block).NextNonce = append((*block).NextNonce, addNextNonce)
		nextNonceIndex = len((*block).NextNonce) - 1
	}
	expectedNonce := (*block).NextNonce[nextNonceIndex].Nonce

	if expectedNonce > (*tx).Info.Nonce {
		fmt.Printf("Replayed transaction %s", tx.Id())
		return false
	} else if expectedNonce < (*tx).Info.Nonce {
		fmt.Printf("Out of order transaction %s", tx.Id())
		return false
	}
	var nextNonce uint32 = expectedNonce + 1
	(*block).NextNonce[nextNonceIndex].Nonce = nextNonce

	var txId string = tx.Id()
	txData := TransactionType{Id: txId, Tx: *tx}
	(*block).Transactions = append((*block).Transactions, txData)

	var senderBalance uint32 = block.BalanceOf((*tx).Info.From)
	senderBalanceIndex := block.FindBalanceIndex((*tx).Info.From)
	(*block).Balances[senderBalanceIndex].Balance = senderBalance - tx.TotalOutput()

	for _, output := range (*tx).Info.Outputs {
		var oldBalance uint32 = (*block).BalanceOf(output.Address)
		oldBalanceId := block.FindBalanceIndex(output.Address)
		if oldBalanceId == -1 {
			balanceAcct := BalanceType{Id: output.Address, Balance: oldBalance + output.Amount}
			(*block).Balances = append((*block).Balances, balanceAcct)
		} else {
			(*block).Balances[oldBalanceId].Balance = oldBalance + output.Amount
		}
	}

	return true
}

func (block *Block) Rerun(prevBlock *Block) bool {

	if prevBlock == nil {
		return false
	}

	// copy Balances from previous block
	block.Balances = make([]BalanceType, 0)
	if prevBlock != nil && (*prevBlock).Balances != nil {
		block.Balances = append(block.Balances, (*prevBlock).Balances...)
	}

	// copy NextNounce from previous block
	block.NextNonce = make([]NextNonceType, 0)
	if prevBlock != nil && (*prevBlock).NextNonce != nil {
		block.NextNonce = append(block.NextNonce, (*prevBlock).NextNonce...)
	}

	// Adding coinbase reward for previous block
	if (*prevBlock).RewardAddr != "" {
		var winnerBalance uint32 = (*prevBlock).BalanceOf((*prevBlock).RewardAddr)
		index := block.FindBalanceIndex((*prevBlock).RewardAddr)

		if index == -1 {
			newBalance := BalanceType{Id: (*prevBlock).RewardAddr, Balance: winnerBalance + prevBlock.CoinbaseReward}
			(*block).Balances = append((*block).Balances, newBalance)
		} else {
			(*block).Balances[index].Balance = winnerBalance + prevBlock.CoinbaseReward
		}
	}

	// Re-enter all transactions
	txMap := make([]TransactionType, len((*block).Transactions))
	copy(txMap, (*block).Transactions)
	(*block).Transactions = make([]TransactionType, 0)
	for _, v := range txMap {
		if !block.AddTransaction(&v.Tx) {
			return false
		}
	}
	return true
}

func (block *Block) BalanceOf(address string) uint32 {
	index := block.FindBalanceIndex(address)
	if index > -1 {
		return block.Balances[index].Balance
	} else {
		return 0
	}
}

func (block *Block) SufficientFund(tx *Transaction) bool {
	var totalOutput uint32 = (*tx).TotalOutput()
	return totalOutput <= (*block).BalanceOf(tx.Info.From)
}

func (block *Block) TotalRewards() uint32 {
	var total uint32 = 0
	for _, v := range (*block).Transactions {
		total += v.Tx.Info.Fee
	}
	total += (*block).CoinbaseReward
	return total
}

func (block *Block) Contains(tx *Transaction) bool {
	index := block.FindTransactionIndex(tx.Id())
	return index > -1
}
