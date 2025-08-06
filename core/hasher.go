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
	return sha256.Sum256(b.HeaderData())
}

var _ Hasher[*Block] = (*BlockHasher)(nil)
