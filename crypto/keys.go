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
		value: ed25519.Sign(k.key, msg),
	}
}

func (k *PrivateKey) PublicKey() *PublicKey {
	// key := make([]byte, kPublicKeyLen)
	// copy(key, k.key[32:])
	key, _ := k.key.Public().(ed25519.PublicKey)
	return &PublicKey{
		key: key,
	}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func (k PublicKey) Bytes() []byte {
	return k.key
}

func (k PublicKey) String() string {
	return "0x" + hex.EncodeToString(k.key)
}

func (k PublicKey) Address() Address {
	return Address{
		value: k.key[len(k.key)-kAddressLen:],
	}
}

type Signature struct {
	value []byte
}

func (s Signature) Bytes() []byte {
	return s.value
}

func (s Signature) String() string {
	return "0x" + hex.EncodeToString(s.value)
}

func (s Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
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
