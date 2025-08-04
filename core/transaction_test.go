package core

import (
	"testing"

	"github.com/matrix-go/block/crypto"
	"github.com/stretchr/testify/require"
)

func TestTransactionSign(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	tx := NewTransaction([]byte("hello"))
	err = tx.Sign(privKey)
	require.NoError(t, err)
	t.Logf("tx signature ===> %+v", tx.Signature)
	t.Logf("tx publicKey ===> %+v", tx.PublicKey)
	err = tx.Verify()
	require.NoError(t, err)
}
