package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	headers   []*Header
	storage   Storage
	validator Validator
}

func NewBlockchain(genesis *Block) (*Blockchain, error) {
	bc := &Blockchain{
		headers:   make([]*Header, 0),
		storage:   NewMemStorage(),
		validator: NewBlockValidator(),
	}
	err := bc.addBlock(genesis)
	return bc, err
}

func (bc *Blockchain) SetValidator(validator Validator) {
	bc.validator = validator
}

func (bc *Blockchain) AddBlock(block *Block) error {

	// validate
	if err := bc.validator.ValidateBlock(bc, block); err != nil {
		return err
	}
	// add block
	if err := bc.addBlock(block); err != nil {
		return err
	}
	return nil
}

// addBlock
// addBlock without validation
func (bc *Blockchain) addBlock(block *Block) error {
	bc.headers = append(bc.headers, block.Header)
	logrus.WithFields(logrus.Fields{
		"height":    block.Height,
		"hash":      block.Hash(NewHeaderHasher()),
		"timestamp": block.Timestamp,
	}).
		Info("add new block")
	return bc.storage.Put(block)
}

func (bc *Blockchain) Height() uint64 {
	return uint64(len(bc.headers) - 1)
}

func (bc *Blockchain) HasBlock(height uint64) bool {
	return bc.Height() >= height // add genesis block && bc.Height() != math.MaxUint64
}

func (bc *Blockchain) GetHeader(height uint64) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height too high")
	}
	return bc.headers[height], nil
}
