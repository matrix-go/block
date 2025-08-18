package network

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestConnectTransport(t *testing.T) {
	trA := NewLocalTransport("a")
	trB := NewLocalTransport("b")
	peerA := NewLocalPeer(trA.Addr(), trA.RpcChan)
	peerB := NewLocalPeer(trB.Addr(), trB.RpcChan)
	err := trA.Connect(peerB)
	require.NoError(t, err)
	err = trB.Connect(peerA)
	require.NoError(t, err)
}

func TestSendMessage(t *testing.T) {
	trA := NewLocalTransport("a")
	trB := NewLocalTransport("b")

	peerA := NewLocalPeer(trA.Addr(), trA.RpcChan)
	peerB := NewLocalPeer(trB.Addr(), trB.RpcChan)
	err := trA.Connect(peerB)
	require.NoError(t, err)
	err = trB.Connect(peerA)
	require.NoError(t, err)

	msg := []byte("hello")
	err = trA.SendMessage(peerB, msg)
	require.NoError(t, err)

	received := <-trB.Consume()
	msgReceived, err := io.ReadAll(received.Payload)
	require.NoError(t, err)
	assert.Equal(t, msg, msgReceived)
}

//func TestLocalTransport_Broadcast(t *testing.T) {
//	trA := NewLocalTransport("a")
//	trB := NewLocalTransport("b")
//	trC := NewLocalTransport("c")
//	err := trA.Connect(trB)
//	require.NoError(t, err)
//	err = trA.Connect(trC)
//	require.NoError(t, err)
//	err = trB.Connect(trA)
//	require.NoError(t, err)
//	err = trB.Connect(trC)
//	require.NoError(t, err)
//	err = trC.Connect(trA)
//	require.NoError(t, err)
//	err = trC.Connect(trB)
//	require.NoError(t, err)
//	require.NoError(t, err)
//	assert.Equal(t, trA, trB.peers[trA.addr])
//	assert.Equal(t, trB, trA.peers[trB.addr])
//
//	// broad cast
//	msg := []byte("hello")
//	err = trA.Broadcast(msg)
//	require.NoError(t, err)
//
//	rpcMsg := <-trB.Consume()
//
//	msgB, err := io.ReadAll(rpcMsg.Payload)
//	require.NoError(t, err)
//	require.Equal(t, msg, msgB)
//	rpcMsg = <-trC.Consume()
//	msgC, err := io.ReadAll(rpcMsg.Payload)
//	require.NoError(t, err)
//	require.Equal(t, msg, msgC)
//
//	time.Sleep(10 * time.Second)
//}
