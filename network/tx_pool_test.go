package network

import (
	"github.com/matrix-go/block/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestTxPool_AddTx(t *testing.T) {
	pool := NewTxPool(10)
	assert.Equal(t, 0, pool.PendingCount())

	tx := core.NewTransaction([]byte("foo"))
	err := pool.Add(tx)
	require.NoError(t, err)
	assert.Equal(t, 1, pool.PendingCount())

	pool.ClearPending()
	assert.Equal(t, 0, pool.PendingCount())
	err = pool.Add(tx)
	require.NoError(t, err)
}

func TestTxPool_SortTransactions(t *testing.T) {
	pool := NewTxPool(10)

	txLen := 1000
	for i := 0; i < txLen; i++ {
		tx := core.NewTransaction([]byte(strconv.Itoa(i)))
		tx.SetFirstSeen(int64(i + 1))
		err := pool.Add(tx)
		require.NoError(t, err)
	}

	assert.Equal(t, txLen, pool.PendingCount())

	txs := pool.Pending()
	for i := 0; i < txLen-1; i++ {
		assert.True(t, txs[i].FirstSeen() <= txs[i+1].FirstSeen())
	}
}
