package core

import (
	"fmt"
	"io"

	"github.com/matrix-go/block/crypto"
)

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Sign(privKey *crypto.PrivateKey) error {
	sig := privKey.Sign(tx.Data)

	tx.From = *privKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return fmt.Errorf("transaction has no signature")
	}
	if tx.Signature.Verify(&tx.From, tx.Data) {
		return nil
	}
	return fmt.Errorf("transaction signature verified failed")
}

func (tx *Transaction) EncodeBinary(w io.Writer) error {
	return nil
}

func (tx *Transaction) DecodeBinary(r io.Reader) error {
	return nil
}
