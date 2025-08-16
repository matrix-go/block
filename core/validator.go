package core

import (
	"errors"
	"fmt"
)

type Validator interface {
	ValidateBlock(bc *Blockchain, block *Block) error
}

type BlockValidator struct {
}

func NewBlockValidator() *BlockValidator {
	return &BlockValidator{}
}

func (b *BlockValidator) ValidateBlock(bc *Blockchain, block *Block) error {
	// height of block
	if bc.HasBlock(block.Height) {
		return ErrBlockAlreadyInBlockchain
	}
	// too high
	if block.Height != bc.Height()+1 {
		return fmt.Errorf("block %s, %w", block.Hash(NewHeaderHasher()), ErrBlockTooHigh)
	}

	// verify block
	if err := block.Verify(); err != nil {
		return err
	}

	header, err := bc.GetHeader(bc.Height())
	if err != nil {
		return err
	}
	hash := NewHeaderHasher().Hash(header)
	if block.PrevHash != hash {
		return ErrBlockPrevHashInvalid
	}
	return nil
}

var _ Validator = (*BlockValidator)(nil)

var (
	ErrBlockTooHigh             = errors.New("block too high")
	ErrBlockAlreadyInBlockchain = errors.New("block already in blockchain")
	ErrBlockPrevHashInvalid     = errors.New("block prev hash invalid")
)
