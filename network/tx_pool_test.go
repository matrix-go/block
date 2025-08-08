package network

import (
	"github.com/matrix-go/block/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"strconv"
	"testing"
)

func TestTxPool_AddTx(t *testing.T) {
	pool := NewTxPool()
	assert.Equal(t, 0, pool.Len())

	tx := core.NewTransaction([]byte("foo"))
	err := pool.AddTx(tx)
	require.NoError(t, err)
	assert.Equal(t, 1, pool.Len())

	pool.Flush()
	assert.Equal(t, 0, pool.Len())
	err = pool.AddTx(tx)
	require.NoError(t, err)
}

func TestTxPool_SortTransactions(t *testing.T) {
	pool := NewTxPool()

	txLen := 1000
	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.Itoa(i)))
		tx.SetFirstSeen(int64(rand.IntN(txLen)))
		err := pool.AddTx(tx)
		require.NoError(t, err)
	}

	assert.Equal(t, txLen, pool.Len())

	txs := pool.Transactions()
	for i := 0; i < txLen-1; i++ {
		assert.True(t, txs[i].FirstSeen() <= txs[i+1].FirstSeen())
	}
}
