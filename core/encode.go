package core

import (
	"encoding/gob"
	"io"
)

type Encoder[T any] interface {
	Encode(t T) error
}

type Decoder[T any] interface {
	Decode(t T) error
}

type TxEncoder struct {
	encoder *gob.Encoder
}

func NewTxEncoder(w io.Writer) *TxEncoder {
	return &TxEncoder{
		encoder: gob.NewEncoder(w),
	}
}

func (enc *TxEncoder) Encode(tx *Transaction) error {
	return enc.encoder.Encode(tx)
}

var _ Encoder[*Transaction] = (*TxEncoder)(nil)

type TxDecoder struct {
	decoder *gob.Decoder
}

func NewTxDecoder(r io.Reader) *TxDecoder {
	return &TxDecoder{
		decoder: gob.NewDecoder(r),
	}
}

func (dec *TxDecoder) Decode(tx *Transaction) error {
	return dec.decoder.Decode(tx)
}

var _ Decoder[*Transaction] = (*TxDecoder)(nil)

type GobBlockEncoder struct {
	encoder *gob.Encoder
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		encoder: gob.NewEncoder(w),
	}
}

func (e *GobBlockEncoder) Encode(b *Block) error {
	return e.encoder.Encode(b)
}

var _ Encoder[*Block] = (*GobBlockEncoder)(nil)

type GobBlockDecoder struct {
	decoder *gob.Decoder
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		decoder: gob.NewDecoder(r),
	}
}

func (e *GobBlockDecoder) Decode(b *Block) error {
	return e.decoder.Decode(b)
}

var _ Decoder[*Block] = (*GobBlockDecoder)(nil)
