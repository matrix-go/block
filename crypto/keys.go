package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

const (
	kPrivateKeyLen = 64
	kPublicKeyLen  = 32
	kSeedLen       = 32
	kAddressLen    = 20
)

var (
	ErrInvalidSeedLength = errors.New("invalid seed length, must be 32 bytes")
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKeyFromString(seedStr string) (*PrivateKey, error) {
	seed, err := hex.DecodeString(seedStr)
	if err != nil {
		return nil, err
	}
	return NewPrivateKeyFromSeed(seed)
}

func NewPrivateKeyFromSeed(seed []byte) (*PrivateKey, error) {
	if len(seed) != kSeedLen {
		return nil, ErrInvalidSeedLength
	}
	// privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}, nil
}

func GeneratePrivateKey() (*PrivateKey, error) {
	seed := make([]byte, kSeedLen)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		return nil, err
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}, nil
}

func (k *PrivateKey) Bytes() []byte {
	return k.key
}

func (k *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{
		Value: ed25519.Sign(k.key, msg),
	}
}

func (k *PrivateKey) PublicKey() *PublicKey {
	// Key := make([]byte, kPublicKeyLen)
	// copy(Key, k.Key[32:])
	key, _ := k.key.Public().(ed25519.PublicKey)
	return &PublicKey{
		Key: key,
	}
}

type PublicKey struct {
	Key ed25519.PublicKey
}

func (k PublicKey) Bytes() []byte {
	return k.Key
}

func (k PublicKey) String() string {
	return "0x" + hex.EncodeToString(k.Key)
}

func (k PublicKey) Address() Address {
	return Address{
		value: k.Key[len(k.Key)-kAddressLen:],
	}
}

type Signature struct {
	Value []byte
}

func (s Signature) Bytes() []byte {
	return s.Value
}

func (s Signature) String() string {
	return "0x" + hex.EncodeToString(s.Value)
}

func (s Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.Key, msg, s.Value)
}

type Address struct {
	value []byte
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

func (a Address) Bytes() []byte {
	return a.value
}
