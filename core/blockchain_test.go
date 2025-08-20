package core

import (
	"github.com/go-kit/log"
	"github.com/matrix-go/block/crypto"
	"os"
	"testing"

	"github.com/matrix-go/block/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newBlockChainWithGenesisBlock(t *testing.T) (bc *Blockchain) {
	genesis := randomBlockWithSignature(0, types.Hash{})
	logger := log.NewLogfmtLogger(os.Stderr)
	bc, err := NewBlockchain(genesis, logger)
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

func TestSendNativeTransfer(t *testing.T) {
	chain := newBlockChainWithGenesisBlock(t)
	bobPrivateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	alicePrivateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	var (
		bobPubKey           = bobPrivateKey.PublicKey()
		alicePubKey         = alicePrivateKey.PublicKey()
		bobAddress          = bobPubKey.Address()
		aliceAddress        = alicePubKey.Address()
		amount       uint64 = 1000
	)

	// initial bob with amount
	err = chain.accountState.CreateAccount(bobAddress)
	require.NoError(t, err)
	err = chain.accountState.AddBalance(bobAddress, amount)
	require.NoError(t, err)

	tx := NewTransaction(nil)
	tx.From = bobPubKey
	tx.To = alicePubKey
	tx.Value = amount
	err = tx.Sign(bobPrivateKey)
	require.NoError(t, err)

	block, err := chain.GetBlock(chain.Height())
	require.NoError(t, err)

	txs := []*Transaction{tx}
	newBlock, err := NewBlockWithPrevHeader(block.Header, txs)
	require.NoError(t, err)

	// mine the block
	minner, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	err = newBlock.Sign(minner)
	require.NoError(t, err)

	// add block
	err = chain.AddBlock(newBlock)
	require.NoError(t, err)

	// assert balance
	bobBalance, err := chain.GetBalance(bobAddress)
	require.NoError(t, err)
	require.Equal(t, bobBalance, uint64(0))
	aliceBalance, err := chain.GetBalance(aliceAddress)
	require.NoError(t, err)
	require.Equal(t, aliceBalance, amount)

}

func TestSendNativeTransferWithHacker(t *testing.T) {
	chain := newBlockChainWithGenesisBlock(t)
	bobPrivateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	alicePrivateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	// hacker
	hackerPrivateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)

	var (
		bobPubKey            = bobPrivateKey.PublicKey()
		alicePubKey          = alicePrivateKey.PublicKey()
		hackerPubKey         = hackerPrivateKey.PublicKey()
		bobAddress           = bobPubKey.Address()
		aliceAddress         = alicePubKey.Address()
		hackerAddress        = hackerPubKey.Address()
		amount        uint64 = 1000
	)

	// initial bob with amount
	err = chain.accountState.CreateAccount(bobAddress)
	require.NoError(t, err)
	err = chain.accountState.AddBalance(bobAddress, amount)
	require.NoError(t, err)

	// initial hacker
	err = chain.accountState.CreateAccount(hackerAddress)
	require.NoError(t, err)

	tx := NewTransaction(nil)
	tx.From = bobPubKey
	tx.To = alicePubKey
	tx.Value = amount
	err = tx.Sign(bobPrivateKey)
	require.NoError(t, err)

	block, err := chain.GetBlock(chain.Height())
	require.NoError(t, err)

	// hacker modify the to value and send the tx
	tx.To = hackerPubKey

	txs := []*Transaction{tx}
	newBlock, err := NewBlockWithPrevHeader(block.Header, txs)
	require.NoError(t, err)

	// mine the block
	minner, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	err = newBlock.Sign(minner)
	require.NoError(t, err)

	// add block
	err = chain.AddBlock(newBlock)
	assert.ErrorIs(t, err, ErrTransactionVerifyFailed)

	// assert balance
	bobBalance, err := chain.GetBalance(bobAddress)
	require.NoError(t, err)
	require.Equal(t, bobBalance, amount)

	_, err = chain.GetBalance(aliceAddress)
	assert.ErrorIs(t, err, ErrAccountNotFound)

	hackerBalance, err := chain.GetBalance(hackerAddress)
	require.NoError(t, err)
	require.Equal(t, hackerBalance, uint64(0))
}
