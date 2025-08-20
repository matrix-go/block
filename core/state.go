package core

import (
	"errors"
	"fmt"
	"github.com/matrix-go/block/types"
	"sync"
)

type AccountState struct {
	lock  sync.RWMutex
	state map[types.Address]uint64
}

func NewAccountState() *AccountState {
	return &AccountState{
		state: make(map[types.Address]uint64),
	}
}

func (s *AccountState) GetBalance(addr types.Address) (balance uint64, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	balance, _ = s.state[addr]
	return
}

func (s *AccountState) AddBalance(to types.Address, amount uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state[to] += amount
	return nil
}

func (s *AccountState) SubBalance(from types.Address, amount uint64) error {
	s.lock.RLock()
	balance := s.state[from]
	if balance < amount {
		s.lock.RUnlock()
		return ErrNotEnoughBalance
	}
	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state[from] -= amount
	return nil
}

func (s *AccountState) Transfer(from types.Address, to types.Address, amount uint64) error {
	if err := s.SubBalance(from, amount); err != nil {
		return err
	}
	return s.AddBalance(to, amount)
}

var (
	ErrNotEnoughBalance = errors.New("not enough balance")
)

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{data: make(map[string][]byte)}
}

func (s *State) Put(k, v []byte) error {
	s.data[string(k)] = v
	return nil
}

func (s *State) Delete(k []byte) error {
	delete(s.data, string(k))
	return nil
}

func (s *State) Get(k []byte) ([]byte, error) {
	if v, ok := s.data[string(k)]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("key not found")
}
