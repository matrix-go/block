package network

import (
	"github.com/stretchr/testify/require"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestTcpTransport_Start(t *testing.T) {
	tr := NewTcpTransport(":8080")
	err := tr.Start()
	require.NoError(t, err)

	// connect
	for i := 0; i < 10; i++ {
		go tcpClientConnect()
	}

	time.Sleep(1 * time.Second)
	err = tr.Stop()
	require.NoError(t, err)

}

func tcpClientConnect() {
	client, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	defer client.Close()
	for i := 0; i < 10; i++ {
		_, err = client.Write([]byte("hello world " + strconv.Itoa(i)))
	}
}
