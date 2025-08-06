package core

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlockchain_AddBlock(t *testing.T) {
	bc := NewBlockchain(nil)
	err := bc.AddBlock(NewBlock(nil, nil))
	require.NoError(t, err)
}
