package core

import (
	"bytes"
	"testing"
	"time"

	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomBlock(height uint64, prevHash types.Hash) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: prevHash,
		Height:        height,
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	txs := []*Transaction{}
	return NewBlock(header, txs)
}

func randomBlockWithSignature(height uint64, prevHash types.Hash) *Block {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	b := randomBlock(height, prevHash)
	tx := randomTxWithSignature()
	b.AddTransaction(tx)
	err = b.Sign(privKey)
	if err != nil {
		panic(err)
	}
	return b
}

func TestBlockSignAndVerify(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	b := randomBlockWithSignature(0, types.Hash{})
	err = b.Sign(privKey)
	require.NoError(t, err)
	t.Logf("block signature ===> %v", b.Signature)
	t.Logf("block validator ===> %v", b.Validator)
	assert.Equal(t, privKey.PublicKey().Address(), b.Validator.Address())
	err = b.Verify()
	require.NoError(t, err)
	b.Height = 100
	err = b.Verify()
	assert.ErrorIs(t, err, ErrBlockVerifyFailed)
}

func TestBlockHeaderEncodeAndDecode(t *testing.T) {
	h := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     uint64(time.Now().UnixMilli()),
		Height:        1,
		Nonce:         15,
	}
	buf := &bytes.Buffer{}
	err := h.EncodeBinary(buf)
	require.NoError(t, err)

	hDecode := &Header{}
	err = hDecode.DecodeBinary(buf)
	require.NoError(t, err)
	assert.Equal(t, h.Version, hDecode.Version)
	assert.Equal(t, h.PrevBlockHash, hDecode.PrevBlockHash)
	assert.Equal(t, h.Timestamp, hDecode.Timestamp)
	assert.Equal(t, h.Height, hDecode.Height)
	assert.Equal(t, h.Nonce, hDecode.Nonce)
}

func TestBlockEncodeAndDecode(t *testing.T) {
	// TODO: encode and decode
	// b := randomBlock(0, types.Hash{})
	// buf := &bytes.Buffer{}
	// encDec := NewBLockEncoderDecoder()
	// err := b.Encode(buf, encDec)
	// require.NoError(t, err)

	// bDecode := randomBlock(0, types.Hash{})
	// err = bDecode.Decode(buf, encDec)
	// require.NoError(t, err)
	// assert.Equal(t, b.Header, bDecode.Header)
	// assert.Equal(t, len(b.Transactions), len(bDecode.Transactions))
	// for idx, tx := range b.Transactions {
	// 	assert.Equal(t, tx.From.Address(), bDecode.Transactions[idx].From.Address())
	// }
	// t.Logf("block ===> %+v", bDecode)
}

func TestBlockHash(t *testing.T) {
	b := randomBlockWithSignature(0, types.Hash{})
	hasher := NewHeaderHasher()
	hash := b.Hash(hasher)
	assert.False(t, hash.IsZero())
	t.Logf("block hash: ====> %v", hash)
}
