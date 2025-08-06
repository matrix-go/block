package core

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"

	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
)

type Header struct {
	Version       uint32
	DataHash      types.Hash
	PrevBlockHash types.Hash
	Timestamp     uint64
	Height        uint64
	Nounce        uint64
}

func (h *Header) Bytes() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	_ = enc.Encode(h)
	return buf.Bytes()
}

func (h *Header) EncodeBinary(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.PrevBlockHash); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Timestamp); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Height); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, &h.Nounce)
}
func (h *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.PrevBlockHash); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Height); err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, &h.Nounce)
}

type Block struct {
	*Header
	Transactions []*Transaction

	// validator that mine the block
	Validator crypto.PublicKey
	Signature *crypto.Signature

	// cached hash of block
	hash types.Hash
}

func NewBlock(header *Header, txs []*Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: txs,
	}
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (b *Block) Sign(privKey *crypto.PrivateKey) error {
	b.Signature = privKey.Sign(b.Header.Bytes())
	b.Validator = *privKey.PublicKey()
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return fmt.Errorf("block has no signature")
	}
	if !b.Signature.Verify(&b.Validator, b.Header.Bytes()) {
		return ErrBlockVerifyFailed
	}
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}
	return nil
}

func (b *Block) Hash(hasher Hasher[*Header]) types.Hash {
	if b.hash.IsZero() {
		b.hash = hasher.Hash(b.Header)
	}
	return b.hash
}

func (b *Block) Encode(w io.Writer, enc Encoder[*Block]) error {
	return enc.Encode(w, b)
}

func (b *Block) Decode(r io.Reader, dec Decoder[*Block]) error {
	return dec.Decode(r, b)
}

var (
	ErrBlockVerifyFailed = errors.New("block valid failed")
)
