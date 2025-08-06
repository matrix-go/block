package core

import (
	"testing"

	"github.com/matrix-go/block/crypto"
	"github.com/stretchr/testify/require"
)

func TestTransactionSignAndVerify(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	tx := NewTransaction([]byte("hello"))
	err = tx.Sign(privKey)
	require.NoError(t, err)
	t.Logf("tx signature ===> %+v", tx.Signature)
	t.Logf("tx publicKey ===> %+v", tx.From)
	err = tx.Verify()
	require.NoError(t, err)
}

func randomTxWithSignature() *Transaction {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	tx := &Transaction{
		Data: []byte("foo"),
	}
	if err := tx.Sign(privKey); err != nil {
		panic(err)
	}
	return tx
}
