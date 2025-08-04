package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectTransport(t *testing.T) {
	trA := NewLocalTransport("a")
	trB := NewLocalTransport("b")
	err := trA.Connect(trB)
	require.NoError(t, err)
	err = trB.Connect(trA)
	require.NoError(t, err)
	assert.Equal(t, trA, trB.peers[trA.addr])
	assert.Equal(t, trB, trA.peers[trB.addr])
}

func TestSendMessage(t *testing.T) {
	trA := NewLocalTransport("a")
	trB := NewLocalTransport("b")
	err := trA.Connect(trB)
	require.NoError(t, err)
	err = trB.Connect(trA)
	require.NoError(t, err)
	assert.Equal(t, trA, trB.peers[trA.addr])
	assert.Equal(t, trB, trA.peers[trB.addr])

	msg := []byte("hello")
	err = trA.SendMessage(trB.addr, msg)
	require.NoError(t, err)

	recevied := <-trB.rpcChan
	assert.Equal(t, msg, recevied.Payload)
}
