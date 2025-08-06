package core

type Validator interface {
	ValidateBlock(block *Block) error
}

type BlockValidator struct {
	bc *Blockchain
}

func NewBlockValidator(bc *Blockchain) *BlockValidator {
	return &BlockValidator{bc}
}

func (b *BlockValidator) ValidateBlock(block *Block) error {

	//
	b.bc.Height()
	block.Height

	return nil
}

var _ Validator = (*BlockValidator)(nil)
