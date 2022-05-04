package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type Block struct {
	PrevBlockHash  []byte
	Target         big.Int
	Proof          uint32
	Balances       map[string]uint32
	NextNonce      map[string]uint32
	Transactions   map[string]Transaction
	TxList         []string
	ChainLength    uint32
	Timestamp      time.Time
	RewardAddr     string
	CoinbaseReward uint32
}

func NewBlock(rewardAddr string, prevBlock *Block, target big.Int, coinbaseReward uint32) *Block {
	var block Block
	block.Target = target
	block.Proof = 0

	block.Balances = make(map[string]uint32)
	if prevBlock != nil && (*prevBlock).Balances != nil {
		for k, v := range (*prevBlock).Balances {
			block.Balances[k] = v
		}
	}

	block.NextNonce = make(map[string]uint32)
	if prevBlock != nil && (*prevBlock).NextNonce != nil {
		for k, v := range (*prevBlock).NextNonce {
			block.NextNonce[k] = v
		}
	}

	block.Transactions = make(map[string]Transaction)
	block.TxList = make([]string, 0)

	block.ChainLength = 0
	if prevBlock != nil {
		block.ChainLength = (*prevBlock).ChainLength + 1
	}

	block.Timestamp = time.Now()
	block.RewardAddr = rewardAddr
	block.CoinbaseReward = coinbaseReward
	return &block
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

func (block *Block) isGenesisBlock() bool {
	if block.ChainLength > 0 {
		return false
	}
	return true
}

func (block *Block) hasValidProof() bool {
	data, err := BlockToBytes(block)
	if err != nil {
		fmt.Printf("Failed to convert Block to Byte array")
		return false
	}
	block_hash := sha256.Sum256(data)
	block_value := big.NewInt(0)
	block_value.SetBytes(block_hash[:])

	if block_value.Cmp(&(*block).Target) >= 0 {
		fmt.Printf("Invalid Proof")
		return false
	}

	return true
}

func (block *Block) AddTransaction(tx Transaction) bool {
	if (*block).contains(&tx) {
		fmt.Printf("Duplicate transaction %s", tx.Id())
		return false
	} else if tx.Sig == nil {
		fmt.Printf("Unsigned transaction %s", tx.Id())
		return false
	} else if tx.VerifySignature() != true {
		fmt.Printf("Invalid signature for transaction %s", tx.Id())
		return false
	} else if (*block).SufficientFund(&tx) != true {
		fmt.Printf("Insufficient fund for transaction %s", tx.Id())
		return false
	}

	var expectedNonce uint32 = 0
	if val, ok := (*block).NextNonce[tx.Info.From]; ok {
		expectedNonce = val
	}

	if expectedNonce > tx.Info.Nonce {
		fmt.Printf("Replayed transaction %s", tx.Id())
		return false
	} else if expectedNonce < tx.Info.Nonce {
		fmt.Printf("Out of order transaction %s", tx.Id())
		return false
	}
	var nextNonce uint32 = expectedNonce + 1
	(*block).NextNonce[tx.Info.From] = nextNonce

	var txId string = tx.Id()
	(*block).Transactions[txId] = tx
	(*block).TxList = append((*block).TxList, txId)
	var senderBalance uint32 = (*block).BalanceOf(tx.Info.From)
	(*block).Balances[tx.Info.From] = senderBalance - tx.TotalOutput()

	for _, output := range tx.Info.Outputs {
		var oldBalance uint32 = (*block).BalanceOf(output.Address)
		(*block).Balances[output.Address] = oldBalance + output.Amount
	}

	return true
}

func (block *Block) Rerun(prevBlock *Block) bool {

	if prevBlock == nil {
		return false
	}

	// copy Balances from previous block
	(*block).Balances = make(map[string]uint32)
	if (*prevBlock).Balances != nil {
		for k, v := range prevBlock.Balances {
			(*block).Balances[k] = v
		}
	} else {
		return false
	}

	// copy NextNounce from previous block
	(*block).NextNonce = make(map[string]uint32)
	if (*prevBlock).NextNonce != nil {
		for k, v := range prevBlock.NextNonce {
			(*block).NextNonce[k] = v
		}
	} else {
		return false
	}

	// Adding coinbase reward for previous block
	if (*prevBlock).RewardAddr != "" {
		var winnerBalance uint32 = (*prevBlock).BalanceOf((*prevBlock).RewardAddr)
		(*block).Balances[(*prevBlock).RewardAddr] = winnerBalance + prevBlock.CoinbaseReward
	}

	// Re-enter all transactions
	txMap := (*block).Transactions
	txList := (*block).TxList
	(*block).Transactions = make(map[string]Transaction)
	(*block).TxList = make([]string, 0)
	if txMap != nil && txList != nil {
		for _, v := range txList {
			if tx, ok := txMap[v]; ok {
				(*block).AddTransaction(tx)
			}
		}
	} else {
		return false
	}

	return true
}

func (block *Block) BalanceOf(address string) uint32 {
	if val, ok := (*block).Balances[address]; ok {
		return val
	} else {
		(*block).Balances[address] = 0
		return 0
	}
}

func (block *Block) SufficientFund(tx *Transaction) bool {
	var totalOutput uint32 = (*tx).TotalOutput()
	if totalOutput >= (*block).BalanceOf(tx.Info.From) {
		return false
	}
	return true
}

func (block *Block) TotalRewards() uint32 {
	var total uint32 = 0
	for _, v := range (*block).Transactions {
		total += v.Info.Fee
	}
	total += (*block).CoinbaseReward
	return total
}

func (block *Block) contains(tx *Transaction) bool {
	if _, ok := (*block).Transactions[(*tx).Id()]; ok {
		return true
	}
	return false
}
