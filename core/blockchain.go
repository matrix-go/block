package core

type Blockchain struct {
	headers   []*Header
	storage   Storage
	validator Validator
}

func NewBlockchain(storage Storage) *Blockchain {
	bc := &Blockchain{
		headers: make([]*Header, 0),
		storage: storage,
	}
	validator := NewBlockValidator(bc)
	bc.validator = validator
	return bc
}

func (bc *Blockchain) SetValidator(validator Validator) {
	bc.validator = validator
}

func (bc *Blockchain) AddBlock(block *Block) error {

	// validate
	if err := bc.validator.ValidateBlock(block); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) Height() uint64 {
	return uint64(len(bc.headers) - 1)
}
