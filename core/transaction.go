package core

import (
	"encoding/gob"
	"errors"
	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
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
	From      *crypto.PublicKey
	To        *crypto.PublicKey
	Value     uint64 // TODO: big.Int
	Signature *crypto.Signature

	// first local node see the tx
	Timestamp int64
	// cached Hash
	Hash types.Hash

	// inner tx
	InnerTx any // one of MintTx and CollectionTx
}

func NewTransaction(data []byte) *Transaction {
	return &Transaction{
		Data: data,
	}
}

func (tx *Transaction) Sign(privateKey *crypto.PrivateKey) error {
	hash := tx.GetHash(NewTransactionHasher())
	sig := privateKey.Sign(hash.Bytes())

	tx.From = privateKey.PublicKey()
	tx.Signature = sig

	return nil
}

func (tx *Transaction) Verify() error {
	if tx.Signature == nil {
		return ErrTransactionNotSigned
	}
	hash := tx.GetHash(NewTransactionHasher())
	if tx.Signature.Verify(tx.From, hash.Bytes()) {
		return nil
	}
	return ErrTransactionVerifyFailed
}

func (tx *Transaction) GetHash(hasher Hasher[*Transaction]) types.Hash {
	return hasher.Hash(tx)
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

var (
	ErrTransactionVerifyFailed = errors.New("transaction verify failed")
	ErrTransactionNotSigned    = errors.New("transaction not signed")
)
