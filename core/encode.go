package core

import "io"

type Encoder[T any] interface {
	Encode(w io.Writer, t T) error
}

type Decoder[T any] interface {
	Decode(r io.Reader, t T) error
}

type BlockEncoderDecoder struct {
}

func NewBLockEncoderDecoder() *BlockEncoderDecoder {
	return &BlockEncoderDecoder{}
}

func (e *BlockEncoderDecoder) Encode(w io.Writer, b *Block) error {
	if err := b.Header.EncodeBinary(w); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if err := tx.EncodeBinary(w); err != nil {
			return err
		}
	}
	return nil
}
func (e *BlockEncoderDecoder) Decode(r io.Reader, b *Block) error {
	if err := b.Header.DecodeBinary(r); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if err := tx.DecodeBinary(r); err != nil {
			return err
		}
	}
	return nil
}

var _ Encoder[*Block] = (*BlockEncoderDecoder)(nil)
var _ Decoder[*Block] = (*BlockEncoderDecoder)(nil)
