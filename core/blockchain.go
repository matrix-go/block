package core

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/matrix-go/block/types"
	"sync"
)

type Blockchain struct {
	logger           log.Logger
	headers          []*Header
	blocks           []*Block
	blockStore       map[types.Hash][]*Block
	transactionStore map[types.Hash][]*Transaction
	storage          Storage
	validator        Validator

	// TODO: make this an interface
	contractState *State

	lock   sync.RWMutex
	txLock sync.RWMutex
}

func NewBlockchain(genesis *Block, logger log.Logger) (*Blockchain, error) {
	bc := &Blockchain{
		logger:           logger,
		headers:          make([]*Header, 0),
		storage:          NewMemStorage(),
		validator:        NewBlockValidator(),
		blockStore:       make(map[types.Hash][]*Block),
		transactionStore: make(map[types.Hash][]*Transaction),
		contractState:    NewState(),
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
	// run vm code
	for _, tx := range block.Transactions {
		bc.logger.Log("msg", "executing code", "len", len(tx.Data), "GetHash", tx.GetHash(NewTransactionHasher()))
		vm := NewVM(tx.Data, bc.contractState)
		if err := vm.Run(); err != nil {
			return err
		}
		fmt.Printf("vm state ======> %+v\n", vm.contractState)
		res := vm.stack.Shift()
		fmt.Printf("vm result ======> %+v\n", res)
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
	hash := NewHeaderHasher().Hash(block.Header)
	bc.lock.Lock()
	bc.headers = append(bc.headers, block.Header)
	bc.blocks = append(bc.blocks, block)
	bc.blockStore[hash] = append(bc.blockStore[hash], block)
	bc.lock.Unlock()
	bc.txLock.Lock()
	defer bc.txLock.Unlock()
	for _, tx := range block.Transactions {
		bc.transactionStore[tx.Hash] = append(bc.transactionStore[tx.Hash], tx)
	}
	bc.logger.Log("msg", "add new block", "height", block.Height, "GetHash", block.GetHash(NewHeaderHasher()), "txLen", len(block.Transactions))
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

func (bc *Blockchain) GetBlock(height uint64) (*Block, error) {
	if height > bc.Height() {
		return nil, fmt.Errorf("given height too high")
	}
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.blocks[height], nil
}

func (bc *Blockchain) GetBlockByHash(hash types.Hash) ([]*Block, error) {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if blocks, ok := bc.blockStore[hash]; ok {
		return blocks, nil
	}
	return nil, fmt.Errorf("block not found")
}

func (bc *Blockchain) Version() uint32 {
	header, _ := bc.GetHeader(bc.Height())
	return header.Version
}

func (bc *Blockchain) GetTransactionByHash(hash types.Hash) ([]*Transaction, error) {
	bc.txLock.RLock()
	defer bc.txLock.RUnlock()
	if transactions, ok := bc.transactionStore[hash]; ok {
		return transactions, nil
	}
	return nil, fmt.Errorf("transaction not found")
}
