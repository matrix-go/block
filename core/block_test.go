package core

import (
	"bytes"
	"testing"
	"time"

	"github.com/matrix-go/block/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	header := Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     uint64(time.Now().UnixMilli()),
		Height:        1,
		Nounce:        15,
	}
	txs := []Transaction{
		{
			Data: []byte("hello"),
		},
	}
	b := NewBlock(header, txs)

	buf := &bytes.Buffer{}
	encDec := NewBLockEncoderDecoder()
	err := b.Encode(buf, encDec)
	require.NoError(t, err)

	bDecode := &Block{}
	err = bDecode.Decode(buf, encDec)
	require.NoError(t, err)
	assert.Equal(t, b.Header, bDecode.Header)
	assert.Equal(t, b.Transactions, bDecode.Transactions)
	t.Logf("block ===> %+v", bDecode)
}

func TestBlockHash(t *testing.T) {
	header := Header{
		Version:       1,
		PrevBlockHash: types.RandomHash(),
		Timestamp:     uint64(time.Now().UnixMilli()),
		Height:        1,
		Nounce:        15,
	}
	txs := make([]Transaction, 0)
	b := NewBlock(header, txs)

	hasher := NewBlockHasher()
	hash := b.Hash(hasher)
	assert.False(t, hash.IsZero())
	t.Logf("block hash: ====> %v", hash)
}
