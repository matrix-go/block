package core

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Blockchain struct {
	headers   []*Header
	storage   Storage
	validator Validator

	lock sync.RWMutex
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
	bc.lock.Lock()
	bc.headers = append(bc.headers, block.Header)
	bc.lock.Unlock()

	logrus.WithFields(logrus.Fields{
		"height":    block.Height,
		"hash":      block.Hash(NewHeaderHasher()),
		"timestamp": block.Timestamp,
	}).
		Info("add new block")
	return bc.storage.Put(block)
}

func (bc *Blockchain) Height() uint64 {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return uint64(len(bc.headers) - 1)
}

func (bc *Blockchain) HasBlock(height uint64) bool {
	return bc.Height() >= height // add genesis block && bc.Height() != math.MaxUint64
}

func (bc *Blockchain) GetHeader(height uint64) (*Header, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height too high")
	}

	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.headers[height], nil
}
