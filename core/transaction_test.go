package core

import (
	"bytes"
	"testing"

	"github.com/matrix-go/block/crypto"
	"github.com/stretchr/testify/require"
)

func TestTransactionSignAndVerify(t *testing.T) {
	privateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	tx := NewTransaction([]byte("hello"))
	err = tx.Sign(privateKey)
	require.NoError(t, err)
	t.Logf("tx signature ===> %+v", tx.Signature)
	t.Logf("tx publicKey ===> %+v", tx.From)
	err = tx.Verify()
	require.NoError(t, err)
}

func TestTxEncodeAndDecode(t *testing.T) {
	originTx := randomTxWithSignature()
	buf := bytes.NewBuffer(nil)
	encoder := NewTxEncoder(buf)
	err := encoder.Encode(originTx)
	require.NoError(t, err)
	var tx Transaction
	decoder := NewTxDecoder(buf)
	err = decoder.Decode(&tx)
	require.NoError(t, err)
	require.Equal(t, *originTx, tx)
}

func TestNFTTransaction(t *testing.T) {
	collectionTx := &CollectionTx{
		Fee:      200,
		Metadata: []byte("this is a test"),
	}
	privateKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	tx := NewTransaction([]byte("hello"))
	tx.InnerTx = collectionTx
	err = tx.Sign(privateKey)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = tx.Encode(NewTxEncoder(&buf))
	require.NoError(t, err)
	var tx2 Transaction
	err = tx2.Decode(NewTxDecoder(&buf))
	require.NoError(t, err)
	require.Equal(t, *tx, tx2)
}

func randomTxWithSignature() *Transaction {
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	tx := &Transaction{
		Data: []byte("foo"),
	}
	if err := tx.Sign(privateKey); err != nil {
		panic(err)
	}
	return tx
}

func TestSendTransactionWithValue(t *testing.T) {
	fromKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	toKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	tx := &Transaction{
		Data:  []byte("foo"),
		From:  fromKey.PublicKey(),
		To:    toKey.PublicKey(),
		Value: 100,
	}
	err = tx.Sign(fromKey)
	require.NoError(t, tx.Sign(toKey))

	var buf bytes.Buffer
	err = tx.Encode(NewTxEncoder(&buf))
	require.NoError(t, err)
	var tx2 Transaction
	err = tx2.Decode(NewTxDecoder(&buf))
	require.NoError(t, err)
	require.Equal(t, *tx, tx2)
}
