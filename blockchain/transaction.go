package blockchain

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Output struct {
	Address string
	Amount  uint32
}

type TransactionInfo struct {
	From    string
	Nonce   uint32
	Pubkey  rsa.PublicKey //256 bits, 32 bytes
	Fee     uint32
	Outputs []Output // slice
	Data    []byte
}

type Transaction struct {
	Info TransactionInfo
	Sig  []byte
}

func NewTransaction(from string, nonce uint32, pubkey *rsa.PublicKey, sig []byte, fee uint32, outputs []Output, data []byte) (*Transaction, error) {
	var tx Transaction
	tx.Info.From = from
	tx.Info.Nonce = nonce
	tx.Info.Pubkey = *pubkey
	tx.Info.Fee = fee
	if len(outputs) == 0 {
		return nil, errors.New("outputs is empty")
	}
	tx.Info.Outputs = make([]Output, len(outputs))
	copy(tx.Info.Outputs, outputs)
	tx.Info.Data = make([]byte, len(data))
	copy(tx.Info.Data, data)

	tx.Sig = make([]byte, len(sig))
	copy(tx.Sig, sig)
	return &tx, nil
}

func TransactionToBytes(tx *Transaction) ([]byte, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func BytesToTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func TransactionInfoToBytes(tx *TransactionInfo) ([]byte, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func BytesToTransactionInfo(data []byte) (*TransactionInfo, error) {
	var txInfo TransactionInfo
	if err := json.Unmarshal(data, &txInfo); err != nil {
		return nil, err
	}
	return &txInfo, nil
}

func (tran *Transaction) ToString() string {
	info := fmt.Sprintf("from : %s\n"+
		"nonce: %d\n"+
		"pubkey: \n\tN: %x\n\tE: %x\n"+
		"fee: %d\n", (*tran).Info.From, (*tran).Info.Nonce,
		(*tran).Info.Pubkey.N, (*tran).Info.Pubkey.E, (*tran).Info.Fee)

	outputs := "outputs: [\n"
	for _, v := range (*tran).Info.Outputs {
		outputs = outputs + fmt.Sprintf("\t{address: %s amount: %d}\n", v.Address, v.Amount)
	}
	outputs = outputs + "]\n"
	info = info + outputs
	data := fmt.Sprintf("data: %s\nsig:%x", hex.EncodeToString(tran.Info.Data), hex.EncodeToString((*tran).Sig))
	info = info + data

	return info
}

func (txInfo *TransactionInfo) GetHash() []byte {
	data, err := TransactionInfoToBytes(txInfo)
	if err != nil {
		return nil
	}
	tx_hash := sha256.Sum256(data)
	return tx_hash[:]
}

func (tx *Transaction) GetHashStr() string {
	data, _ := TransactionToBytes(tx)
	hashed := sha256.Sum256(data)
	return hex.EncodeToString(hashed[:])
}

func (tx *Transaction) Id() string {
	return tx.GetHashStr()
}

func (tx *Transaction) Sign(privKey *rsa.PrivateKey) []byte {
	rng := rand.Reader
	hashed := (&(*tx).Info).GetHash()
	signature, err := rsa.SignPKCS1v15(rng, privKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return nil
	}
	tx.Sig = make([]byte, len(signature))
	copy(tx.Sig, signature)
	return signature
}

func (tx *Transaction) VerifySignature() bool {
	hashed := (&(*tx).Info).GetHash()
	err := rsa.VerifyPKCS1v15(&tx.Info.Pubkey, crypto.SHA256, hashed[:], (*tx).Sig)
	return err == nil
}

func (tx *Transaction) TotalOutput() uint32 {
	var amount uint32 = 0
	for _, v := range (*tx).Info.Outputs {
		amount += v.Amount
	}
	amount += (*tx).Info.Fee
	return amount
}
