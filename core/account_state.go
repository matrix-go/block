package core

import (
	"errors"
	"fmt"
	"github.com/matrix-go/block/types"
	"sync"
)

type Account struct {
	Address types.Address
	Balance uint64
}

func (a Account) String() string {
	return fmt.Sprintf("address=%+v, balance=%d", a.Address, a.Balance)
}

type AccountState struct {
	lock  sync.RWMutex
	state map[types.Address]*Account
}

func NewAccountState() *AccountState {
	return &AccountState{
		state: make(map[types.Address]*Account),
	}
}

func (s *AccountState) CreateAccount(addr types.Address) error {
	_, err := s.GetAccount(addr)
	if err == nil {
		return ErrAlreadyExists
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state[addr] = &Account{
		Address: addr,
	}
	return nil
}

func (s *AccountState) GetAccount(addr types.Address) (account *Account, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var exists bool
	account, exists = s.state[addr]
	if !exists {
		err = ErrAccountNotFound
	}
	return account, err
}

func (s *AccountState) GetBalance(addr types.Address) (balance uint64, err error) {
	account, err := s.GetAccount(addr)
	if err != nil {
		return balance, err
	}
	return account.Balance, nil
}

func (s *AccountState) AddBalance(to types.Address, amount uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.state[to]; ok {
		s.state[to].Balance += amount
	} else {
		s.state[to] = &Account{
			Address: to,
			Balance: amount,
		}
	}
	return nil
}

func (s *AccountState) SubBalance(from types.Address, amount uint64) error {
	account, err := s.GetAccount(from)
	if err != nil {
		return err
	}
	if account.Balance < amount {
		return ErrInsufficientBalance
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state[from].Balance -= amount
	return nil
}

func (s *AccountState) Transfer(from types.Address, to types.Address, amount uint64) error {
	account, err := s.GetAccount(from)
	if err != nil {
		return err
	}
	if account.Balance < amount {
		return ErrInsufficientBalance
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.state[from].Balance -= amount
	if _, exists := s.state[to]; !exists {
		s.state[to] = &Account{
			Address: to,
		}
	}
	s.state[to].Balance += amount
	return nil
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountNotFound     = errors.New("account not found")
	ErrAlreadyExists       = errors.New("account already exists")
)
