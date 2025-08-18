package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"time"

	"github.com/matrix-go/block/crypto"
	"github.com/matrix-go/block/types"
)

type Header struct {
	Version   uint32
	DataHash  types.Hash
	PrevHash  types.Hash
	Timestamp uint64
	Height    uint64
	Nonce     uint64
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
	if err := binary.Write(w, binary.LittleEndian, &h.PrevHash); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Timestamp); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h.Height); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, &h.Nonce)
}
func (h *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.PrevHash); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Height); err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, &h.Nonce)
}

type Block struct {
	*Header
	Transactions []*Transaction

	// validator that mine the block
	Validator crypto.PublicKey
	Signature *crypto.Signature

	// cached Hash of block
	Hash types.Hash
}

func NewBlock(header *Header, txs []*Transaction) *Block {
	return &Block{
		Header:       header,
		Transactions: txs,
	}
}

func NewBlockWithPrevHeader(prevHeader *Header, txs []*Transaction) (*Block, error) {
	dataHash, err := CalculateDataHash(txs)
	if err != nil {
		return nil, err
	}
	header := &Header{
		Version:   prevHeader.Version,
		DataHash:  dataHash,
		PrevHash:  NewHeaderHasher().Hash(prevHeader),
		Timestamp: uint64(time.Now().UnixNano()),
		Height:    prevHeader.Height + 1,
		Nonce:     prevHeader.Nonce,
	}
	return NewBlock(header, txs), nil
}

func (b *Block) AddTransaction(tx *Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

func (b *Block) Sign(privateKey *crypto.PrivateKey) error {
	b.Signature = privateKey.Sign(b.Header.Bytes())
	b.Validator = *privateKey.PublicKey()
	return nil
}

func (b *Block) Verify() error {
	if b.Signature == nil {
		return ErrorBlockHasNoSig
	}
	if !b.Signature.Verify(&b.Validator, b.Header.Bytes()) {
		return ErrBlockVerifyFailed
	}
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}

	dataHash, err := CalculateDataHash(b.Transactions)
	if err != nil {
		return err
	}
	if b.DataHash != dataHash {
		return ErrBlockInvalidHash
	}
	return nil
}

func (b *Block) GetHash(hasher Hasher[*Header]) types.Hash {
	if b.Hash.IsZero() {
		b.Hash = hasher.Hash(b.Header)
	}
	return b.Hash
}

func (b *Block) Encode(enc Encoder[*Block]) error {
	return enc.Encode(b)
}

func (b *Block) Decode(dec Decoder[*Block]) error {
	return dec.Decode(b)
}

func CalculateDataHash(txs []*Transaction) (hash types.Hash, err error) {
	var buf bytes.Buffer
	for _, tx := range txs {
		if err = tx.Encode(NewTxEncoder(&buf)); err != nil {
			return
		}
	}
	hash = sha256.Sum256(buf.Bytes())
	return
}

var (
	ErrBlockVerifyFailed = errors.New("block valid failed")
	ErrBlockInvalidHash  = errors.New("block has an invalid GetHash")
	ErrorBlockHasNoSig   = errors.New("block has no signature")
)
