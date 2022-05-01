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

type Transaction struct {
	From    string
	Nonce   uint32
	Pubkey  rsa.PublicKey //256 bits, 32 bytes
	Sig     []byte
	Fee     uint32
	Outputs []Output // slice
	Data    []byte
}

func NewTransaction(from string, nonce uint32, pubkey rsa.PublicKey, sig []byte, fee uint32, outputs []Output, data []byte) (*Transaction, error) {
	var transaction Transaction
	transaction.From = from
	transaction.Nonce = nonce
	transaction.Pubkey = pubkey
	transaction.Sig = make([]byte, len(sig))
	copy(transaction.Sig, sig)
	transaction.Fee = fee
	if len(outputs) == 0 {
		return nil, errors.New("outputs is empty")
	}
	transaction.Outputs = make([]Output, len(outputs))
	copy(transaction.Outputs, outputs)
	transaction.Data = make([]byte, len(data))
	copy(transaction.Data, data)
	return &transaction, nil
}

func TransactionToBytes(tran *Transaction) ([]byte, error) {
	data, err := json.Marshal(tran)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func BytesToTransaction(data []byte) (*Transaction, error) {
	var tran1 Transaction
	if err := json.Unmarshal(data, &tran1); err != nil {
		return nil, err
	}
	return &tran1, nil
}

func (tran *Transaction) ToString() string {
	info := fmt.Sprintf("from : %s\n"+
		"nonce: %d\n"+
		"pubkey: \n\tN: %x\n\tE: %x\n"+
		"sig: %s\n"+
		"fee: %d\n", (*tran).From, (*tran).Nonce, (*tran).Pubkey.N, (*tran).Pubkey.E, hex.EncodeToString((*tran).Sig[:]), (*tran).Fee)

	outputs := "outputs: [\n"
	for _, v := range (*tran).Outputs {
		outputs = outputs + fmt.Sprintf("{address: %s amount: %d}\n", v.Address, v.Amount)
	}
	outputs = outputs + "]\n"
	info = info + outputs
	data := fmt.Sprintf("data: %s\n", hex.EncodeToString(tran.Data))
	info = info + data

	return info
}

func (tx *Transaction) GetHash() []byte {
	data, err := TransactionToBytes(tx)
	if err != nil {
		return nil
	}
	tx_hash := sha256.Sum256(data)
	return tx_hash[:]
}

func (tx *Transaction) Sign(privKey *rsa.PrivateKey) []byte {
	rng := rand.Reader
	hashed := tx.GetHash()
	signature, err := rsa.SignPKCS1v15(rng, privKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return nil
	}
	copy(tx.Sig, signature)
	return signature
}

func (tx *Transaction) VerifySignature(privKey *rsa.PrivateKey, signature []byte) bool {
	hashed := tx.GetHash()

	err := rsa.VerifyPKCS1v15(&privKey.PublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		return false
	}
	return true

}

func (tx *Transaction) TotalOutput() uint32 {
	var amount uint32 = 0
	for _, v := range (*tx).Outputs {
		amount += v.Amount
	}
	return amount
}
