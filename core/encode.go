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
