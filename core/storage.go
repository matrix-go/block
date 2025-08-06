package core

type Storage interface {
	Put(b *Block) error
}

type MemStorage struct {
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}

// Put implements Storage.
func (m *MemStorage) Put(b *Block) error {
	return nil
}

var _ Storage = (*MemStorage)(nil)
