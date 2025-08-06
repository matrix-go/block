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

func randomBlock(height uint64) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Height:        height,
		Timestamp:     uint64(time.Now().UnixNano()),
	}
	txs := []Transaction{
		{
			Data: []byte("foo"),
		},
	}
	return NewBlock(header, txs)
}

func TestBlockSignAndVerify(t *testing.T) {
	privKey, err := crypto.GeneratePrivateKey()
	require.NoError(t, err)
	b := randomBlock(0)
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
		Nounce:        15,
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
	assert.Equal(t, h.Nounce, hDecode.Nounce)
}

func TestBlockEncodeAndDecode(t *testing.T) {
	b := randomBlock(0)
	buf := &bytes.Buffer{}
	encDec := NewBLockEncoderDecoder()
	err := b.Encode(buf, encDec)
	require.NoError(t, err)

	bDecode := randomBlock(0)
	err = bDecode.Decode(buf, encDec)
	require.NoError(t, err)
	assert.Equal(t, b.Header, bDecode.Header)
	assert.Equal(t, b.Transactions, bDecode.Transactions)
	t.Logf("block ===> %+v", bDecode)
}

func TestBlockHash(t *testing.T) {
	b := randomBlock(0)
	hasher := NewBlockHasher()
	hash := b.Hash(hasher)
	assert.False(t, hash.IsZero())
	t.Logf("block hash: ====> %v", hash)
}
