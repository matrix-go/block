package network

import (
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/types"
	"sync"
)

type TxPool struct {
	all     *TxSortedMap
	pending *TxSortedMap

	// the max length of the mempool of transactions
	// when the pool is full we will prune the oldest transaction
	maxLength int
}

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		all:       NewSortedMap(),
		pending:   NewSortedMap(),
		maxLength: maxLength,
	}
}

func (p *TxPool) Add(tx *core.Transaction) error {

	if p.all.Count() == p.maxLength {
		oldest := p.all.First()
		p.all.Remove(oldest.GetHash(core.NewTransactionHasher()))
	}
	txHash := tx.GetHash(core.NewTransactionHasher())
	if !p.all.Contains(txHash) {
		p.all.Add(tx)
		p.pending.Add(tx)
	}
	return nil
}

func (p *TxPool) Contains(hash types.Hash) bool {
	return p.all.Contains(hash)
}

func (p *TxPool) Pending() []*core.Transaction {
	return p.pending.txs.Data
}

func (p *TxPool) ClearPending() {
	p.pending.txs.Clear()
}

func (p *TxPool) PendingCount() int {
	return p.pending.Count()
}

type TxSortedMap struct {
	lock   sync.RWMutex
	lookup map[types.Hash]*core.Transaction
	txs    *types.List[*core.Transaction]
}

func NewSortedMap() *TxSortedMap {
	return &TxSortedMap{
		lookup: make(map[types.Hash]*core.Transaction),
		txs:    types.NewList[*core.Transaction](),
	}
}

func (t *TxSortedMap) First() *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()
	first := t.txs.Get(0)
	return t.lookup[first.GetHash(core.NewTransactionHasher())]
}

func (t *TxSortedMap) Remove(hash types.Hash) {
	if !t.Contains(hash) {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.lookup, hash)
	tx := t.lookup[hash]
	t.txs.Remove(tx)
}

func (t *TxSortedMap) Add(tx *core.Transaction) {
	hash := tx.GetHash(core.NewTransactionHasher())
	if contains := t.Contains(hash); contains {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	t.lookup[hash] = tx
	t.txs.Insert(tx)
}

func (t *TxSortedMap) Contains(hash types.Hash) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	_, exists := t.lookup[hash]
	return exists
}

func (t *TxSortedMap) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.txs.Count()
}
