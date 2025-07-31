package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	require.NoError(t, err)
	assert.Equal(t, kPrivateKeyLen, len(privKey.Bytes()))
	pubKey := privKey.PublicKey()
	assert.Equal(t, kPublicKeyLen, len(pubKey.Bytes()))
}

func TestVerifySignature(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	require.NoError(t, err)
	msg := []byte("Hello")
	sig := privKey.Sign(msg)

	// test with valid key
	verified := sig.Verify(privKey.PublicKey(), msg)
	assert.True(t, verified)

	// test with invalid msg
	assert.False(t, sig.Verify(privKey.PublicKey(), []byte("world")))

	// test with invalid key
	privKey, err = GeneratePrivateKey()
	require.NoError(t, err)
	assert.False(t, sig.Verify(privKey.PublicKey(), msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey, err := GeneratePrivateKey()
	require.NoError(t, err)
	pubKey := privKey.PublicKey()
	addr := pubKey.Address()
	assert.Equal(t, len(addr.Bytes()), kAddressLen)
	t.Logf("address: => 0x%v", addr)
}
