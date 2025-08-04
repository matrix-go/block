package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Hash [32]uint8

func (h Hash) IsZero() bool {
	for _, b := range h {
		if b != 0 {
			return false
		}
	}
	return true
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func HashFromBytes(b []byte) Hash {
	if len(b) != 32 {
		panic(fmt.Sprintf("invalid provided bytes length, it should be 32 but actualy it is %d.", len(b)))
	}
	res := [32]uint8{}
	copy(res[:], b)
	return res
}

func RandomBytes(size uint) []byte {
	token := make([]byte, size)
	_, _ = rand.Read(token)
	return token
}

func RandomHash() Hash {
	return HashFromBytes(RandomBytes(32))
}
