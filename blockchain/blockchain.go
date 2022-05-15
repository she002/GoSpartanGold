package blockchain

import (
	"errors"
)

// Network message constants
const MISSING_BLOCK string = "MISSING_BLOCK"
const POST_TRANSACTION string = "POST_TRANSACTION"
const PROOF_FOUND string = "PROOF_FOUND"
const START_MINING string = "START_MINING"

// Constants for mining
const NUM_ROUNDS_MINING uint32 = 2000

// Constants related to proof-of-work target
const POW_BASE_TARGET_STR string = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
const POW_LEADING_ZEROES uint32 = 15

// Constants for mining rewards and default transaction fees
const COINBASE_AMT_ALLOWED uint32 = 25
const DEFAULT_TX_FEE uint32 = 1

// If a block is 6 blocks older than the current block, it is considered
// confirmed, for no better reason than that is what Bitcoin does.
// Note that the genesis block is always considered to be confirmed.
const CONFIRMED_DEPTH uint32 = 6

type BlockchainConfig struct {
	coinbaseAmount uint32
	defaultTxFee   uint32
	confirmedDepth uint32
}

func MakeGenesisDefault(startingBalances map[string]uint32) (*Block, BlockchainConfig, error) {
	return MakeGenesis(POW_LEADING_ZEROES, COINBASE_AMT_ALLOWED, DEFAULT_TX_FEE, CONFIRMED_DEPTH, startingBalances)
}

func MakeGenesis(leading_zeros uint32, coinbase_amt uint32, tx_fee uint32, confirmed_depth uint32, starting_balances map[string]uint32) (*Block, BlockchainConfig, error) {
	var newconfig BlockchainConfig
	newconfig.coinbaseAmount = coinbase_amt
	newconfig.confirmedDepth = confirmed_depth
	newconfig.defaultTxFee = tx_fee

	if starting_balances == nil {
		return nil, newconfig, errors.New("makeGenesis(...): starting_balances cannot be nil")
	}

	target := CalculateTarget(leading_zeros)
	newblock := NewBlock("", nil, target, coinbase_amt)

	for k, v := range starting_balances {
		newBalance := BalanceType{Id: k, Balance: v}
		(*newblock).Balances = append((*newblock).Balances, newBalance)
	}

	return newblock, newconfig, nil
}
