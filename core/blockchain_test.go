package core

import (
	"github.com/go-kit/log"
	"os"
	"testing"

	"github.com/matrix-go/block/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBlockChainWithGenesisBlock(t *testing.T) (bc *Blockchain) {
	b := randomBlockWithSignature(0, types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	state := NewAccountState()
	bc, err := NewBlockchain(b, state, logger)
	require.NoError(t, err)
	return bc
}

func getPreviousBlockHash(t *testing.T, bc *Blockchain, height uint64) types.Hash {
	header, err := bc.GetHeader(height)
	require.NoError(t, err)
	return NewHeaderHasher().Hash(header)
}

func TestBlockchain_AddBlock(t *testing.T) {
	bc := newBlockChainWithGenesisBlock(t)

	for i := 0; i < 100; i++ {
		prevHash := getPreviousBlockHash(t, bc, bc.Height())
		b := randomBlockWithSignature(uint64(i+1), prevHash)
		err := bc.AddBlock(b)
		require.NoError(t, err)
	}
	for _, header := range bc.headers {
		t.Logf("block headers ====> %+v", header)
	}
	t.Logf("block height ====> %v", bc.Height())
}

func TestHasBlock(t *testing.T) {
	bc := newBlockChainWithGenesisBlock(t)
	assert.True(t, bc.HasBlock(0))
	assert.False(t, bc.HasBlock(1))
	assert.False(t, bc.HasBlock(100))
}

func TestBlockTooHigh(t *testing.T) {
	bc := newBlockChainWithGenesisBlock(t)
	prevHash := getPreviousBlockHash(t, bc, bc.Height())
	err := bc.AddBlock(randomBlockWithSignature(3, prevHash))
	t.Logf("got err: %v", err)
	assert.ErrorIs(t, err, ErrBlockTooHigh)
}
