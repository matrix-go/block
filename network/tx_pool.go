package network

import (
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/types"
)

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
	txSorter     TxSorter
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[types.Hash]*core.Transaction),
		txSorter:     NewTxSorter(),
	}
}

func (p *TxPool) Transactions() []*core.Transaction {
	return p.txSorter.SortTransactions(p.transactions)
}

func (p *TxPool) AddTx(tx *core.Transaction) error {
	hasher := core.NewTransactionHasher()
	hash := tx.Hash(hasher)
	p.transactions[hash] = tx
	return nil
}

func (p *TxPool) HasTx(hash types.Hash) bool {
	_, exists := p.transactions[hash]
	return exists
}

func (p *TxPool) Len() int {
	return len(p.transactions)
}

func (p *TxPool) Flush() {
	p.transactions = make(map[types.Hash]*core.Transaction)
}
