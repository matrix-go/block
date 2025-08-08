package core

import (
	"crypto/sha256"

	"github.com/matrix-go/block/types"
)

type Hasher[T any] interface {
	Hash(T) types.Hash
}

type BlockHasher struct {
}

func NewBlockHasher() *BlockHasher {
	return &BlockHasher{}
}

func (h *BlockHasher) Hash(b *Block) types.Hash {
	return sha256.Sum256(b.Header.Bytes())
}

var _ Hasher[*Block] = (*BlockHasher)(nil)

type HeaderHasher struct {
}

func NewHeaderHasher() *HeaderHasher {
	return &HeaderHasher{}
}

func (h *HeaderHasher) Hash(header *Header) types.Hash {
	return sha256.Sum256(header.Bytes())
}

var _ Hasher[*Header] = (*HeaderHasher)(nil)

type TransactionHasher struct {
}

func NewTransactionHasher() *TransactionHasher {
	return &TransactionHasher{}
}

func (h *TransactionHasher) Hash(tx *Transaction) types.Hash {
	return sha256.Sum256(tx.Data)
}

var _ Hasher[*Transaction] = (*TransactionHasher)(nil)
