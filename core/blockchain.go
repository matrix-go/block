package core

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
	"sync"
)

type Blockchain struct {
	logger           log.Logger
	headers          []*Header
	blocks           []*Block
	blockStore       map[types.Hash][]*Block
	transactionStore map[types.Hash][]*Transaction
	collectionStore  map[types.Hash]*CollectionTx
	mintStore        map[types.Hash]*MintTx
	storage          Storage
	validator        Validator

	// TODO: make this an interface
	contractState *State
	accountState  *AccountState

	lock     sync.RWMutex
	txLock   sync.RWMutex
	colLock  sync.RWMutex
	mintLock sync.RWMutex
}

func NewBlockchain(genesis *Block, logger log.Logger) (bc *Blockchain, err error) {
	// TODO: read state from disk db
	accountState := NewAccountState()
	contractState := NewState()

	coinbase := crypto.PublicKey{}
	if err := accountState.CreateAccount(coinbase.Address()); err != nil {
		return nil, err
	}

	bc = &Blockchain{
		logger:           logger,
		headers:          make([]*Header, 0),
		storage:          NewMemStorage(),
		validator:        NewBlockValidator(),
		blockStore:       make(map[types.Hash][]*Block),
		transactionStore: make(map[types.Hash][]*Transaction),
		collectionStore:  make(map[types.Hash]*CollectionTx),
		mintStore:        make(map[types.Hash]*MintTx),
		contractState:    contractState,
		accountState:     accountState,
	}

	if genesis != nil {
		err = bc.addBlock(genesis)
		if err == nil && genesis.Validator != nil {
			coinbaseAccount, err := bc.accountState.GetAccount(coinbase.Address())
			if err != nil {
				return nil, err
			}
			if err = bc.accountState.Transfer(coinbaseAccount.Address, genesis.Validator.Address(), coinbaseAccount.Balance); err != nil {
				return nil, err
			}
		}
	}
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

func (bc *Blockchain) handleNativeTransaction(tx *Transaction) error {
	fmt.Printf("======> %s is going to send %d coin to %s\n", tx.From, tx.Value, tx.To)
	if tx.From.String() == "0x996fb92427ae41e4649b934ca495991b7852b855" {
		return bc.accountState.AddBalance(tx.To.Address(), tx.Value)
	}
	return bc.accountState.Transfer(tx.From.Address(), tx.To.Address(), tx.Value)
}

func (bc *Blockchain) handleNativeNFT(tx *Transaction) error {
	switch innerTx := tx.InnerTx.(type) {
	case *CollectionTx:
		fmt.Printf("tx.InnerTx ======> %+v\n", *innerTx)
		hash := tx.GetHash(NewTransactionHasher())
		bc.colLock.RLock()
		_, exists := bc.collectionStore[hash]
		bc.colLock.RUnlock()
		if exists {
			return fmt.Errorf("collection already exists")
		}
		bc.colLock.Lock()
		bc.collectionStore[hash] = innerTx
		bc.colLock.Unlock()
	case *MintTx:
		bc.colLock.RLock()
		collection, exists := bc.collectionStore[innerTx.Collection]
		bc.colLock.RUnlock()
		if !exists {
			return fmt.Errorf("collection does not exist")
		}
		_ = collection
		hash := tx.GetHash(NewTransactionHasher())
		bc.mintLock.Lock()
		bc.mintStore[hash] = innerTx
		bc.mintLock.Unlock()
		fmt.Printf("tx.InnerTx mint collection ======> %+v\n", *innerTx)
	default:
		return fmt.Errorf("invalid transaction type: %v", innerTx)
	}
	return nil
}

// addBlock
// addBlock without validation
func (bc *Blockchain) addBlock(block *Block) error {

	// run transaction code
	for _, tx := range block.Transactions {
		// handle contract with vm
		if len(tx.Data) > 0 {
			bc.logger.Log("msg", "executing code", "len", len(tx.Data), "Hash", tx.GetHash(NewTransactionHasher()))
			vm := NewVM(tx.Data, bc.contractState)
			if err := vm.Run(); err != nil {
				return err
			}
			fmt.Printf("vm state ======> %+v\n", vm.contractState)
			res := vm.stack.Shift()
			fmt.Printf("vm result ======> %+v\n", res)
		}

		// handle inner transaction
		if tx.InnerTx != nil {
			if err := bc.handleNativeNFT(tx); err != nil {
				return err
			}
		}
		// handle native transaction
		if tx.Value > 0 {
			if err := bc.handleNativeTransaction(tx); err != nil {
				return err
			}
			fmt.Printf("====== ACCOUNT STATE ====== \n")
			fmt.Printf("%+v \n", bc.accountState.state)
			fmt.Printf("====== ACCOUNT STATE ====== \n")
		}
	}

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
	bc.logger.Log("msg", "add new block", "height", block.Height, "Hash", block.GetHash(NewHeaderHasher()), "txLen", len(block.Transactions))
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

func (bc *Blockchain) GetBalance(addr types.Address) (uint64, error) {
	return bc.accountState.GetBalance(addr)
}
