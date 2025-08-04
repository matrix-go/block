package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashFromBytes(t *testing.T) {
	b, err := hex.DecodeString("FFFF00E1CDFF00E10E1CDFF010E1CDFFAF00E1CDFF00E10E1CDFF010E1CDFFAE")
	require.NoError(t, err)
	hash := HashFromBytes(b)
	t.Logf("hash ======> %X", hash)
}
