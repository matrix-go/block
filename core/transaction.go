package core

import (
	"encoding/gob"
	"fmt"
	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
)

type InnerTxType byte

const (
	InnerTxTypeCollection InnerTxType = iota
	InnerTxTypeMint
)

type CollectionTx struct {
	Fee      int64
	Metadata []byte
}

type MintTx struct {
	Fee             int64
	Metadata        []byte
	NFT             types.Hash
	Collection      types.Hash
	CollectionOwner crypto.PublicKey
}

type Transaction struct {
	Data      []byte
	From      crypto.PublicKey
	Signature *crypto.Signature

	// first local node see the tx
	Timestamp int64
	// cached Hash
	Hash types.Hash

	// inner tx
	InnerType InnerTxType
	InnerTx   any // one of MintTx and CollectionTx
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Sign(privateKey *crypto.PrivateKey) error {
	sig := privateKey.Sign(tx.Data)

	tx.From = *privateKey.PublicKey()
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

func (tx *Transaction) GetHash(hasher Hasher[*Transaction]) types.Hash {
	if tx.Hash.IsZero() {
		tx.Hash = hasher.Hash(tx)
	}
	return tx.Hash
}

func (tx *Transaction) Encode(enc Encoder[*Transaction]) error {
	return enc.Encode(tx)
}

func (tx *Transaction) Decode(dec Decoder[*Transaction]) error {
	return dec.Decode(tx)
}

func (tx *Transaction) FirstSeen() int64 {
	return tx.Timestamp
}

func (tx *Transaction) SetFirstSeen(t int64) {
	tx.Timestamp = t
}

func init() {
	gob.Register(&CollectionTx{})
	gob.Register(&MintTx{})
}
