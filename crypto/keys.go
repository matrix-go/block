package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/matrix-go/block/types"
	"io"
	"strings"
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

func NewPrivateKeyFromReader(r io.Reader) (*PrivateKey, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	key := buf.Bytes()
	if len(key) != kPrivateKeyLen {
		return nil, ErrInvalidSeedLength
	}
	return &PrivateKey{
		key: key,
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
	return "0x" + k.Address().String()
}

func (k PublicKey) Address() types.Address {
	h := sha256.Sum256(k.Key)
	return types.AddressFromBytes(h[len(h)-20:])
}

func (k *PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *PublicKey) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	key, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	k.Key = key
	return nil
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

func (s *Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Signature) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	val, err := hex.DecodeString(str)
	if err != nil {
		return err
	}
	s.Value = val
	return nil
}

//type Address struct {
//	value []byte
//}
//
//func (a Address) String() string {
//	return hex.EncodeToString(a.value)
//}
//
//func (a Address) Bytes() []byte {
//	return a.value
//}
