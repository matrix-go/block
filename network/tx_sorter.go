package network

import (
	"github.com/matrix-go/block/core"
	"github.com/matrix-go/block/types"
	"sort"
)

type TxSorter interface {
	SortTransactions(map[types.Hash]*core.Transaction) []*core.Transaction
}

type txSorter struct {
	txs []*core.Transaction
}

func (t *txSorter) SortTransactions(txs map[types.Hash]*core.Transaction) []*core.Transaction {
	t.txs = make([]*core.Transaction, 0, len(txs))
	for _, tx := range txs {
		t.txs = append(t.txs, tx)
	}
	sort.Sort(t)
	return t.txs
}

func (t *txSorter) Len() int {
	return len(t.txs)
}

func (t *txSorter) Less(i, j int) bool {
	return t.txs[i].FirstSeen() < t.txs[j].FirstSeen()
}

func (t *txSorter) Swap(i, j int) {
	t.txs[i], t.txs[j] = t.txs[j], t.txs[i]
}

func NewTxSorter() *txSorter {
	return &txSorter{}
}

var _ TxSorter = (*txSorter)(nil)
var _ sort.Interface = (*txSorter)(nil)
