package core

import (
	"github.com/matrix-go/block/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountState_TransferNotEnoughBalance(t *testing.T) {
	state := NewAccountState()
	fromKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err, "Error generating private key")
	toKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err, "Error generating private key")

	var amount uint64 = 1000
	err = state.Transfer(fromKey.PublicKey().Address(), toKey.PublicKey().Address(), amount)
	assert.ErrorIs(t, err, ErrNotEnoughBalance)
}

func TestAccountState_TransferSuccess(t *testing.T) {
	state := NewAccountState()
	fromKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err, "Error generating private key")
	toKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err, "Error generating private key")

	var amount uint64 = 1000

	err = state.AddBalance(fromKey.PublicKey().Address(), amount)
	require.NoError(t, err, "Error adding balance")
	balance, err := state.GetBalance(fromKey.PublicKey().Address())
	require.NoError(t, err, "Error getting balance")
	assert.Equal(t, amount, balance)
	err = state.Transfer(fromKey.PublicKey().Address(), toKey.PublicKey().Address(), amount)
	assert.NoError(t, err, "Error transferring from")
	balance, err = state.GetBalance(toKey.PublicKey().Address())
	assert.NoError(t, err, "Error getting balance")
	assert.Equal(t, amount, balance)

}
