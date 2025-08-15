package core

import (
	"fmt"
	"github.com/matrix-go/block/util"
)

type Stack struct {
	data []any
	sp   int // stack point
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) Push(v any) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) Pop() any {
	ret := s.data[s.sp-1]
	s.data = append(s.data[:s.sp-1], s.data[s.sp:]...)
	s.sp--
	return ret
}

func (s *Stack) Shift() any {
	if s.sp == 0 {
		return nil
	}
	ret := s.data[0]
	s.data = append(s.data[1:], nil)
	s.sp--
	return ret
}

type VM struct {
	data          []byte
	ip            int // instruction pointer
	stack         *Stack
	contractState *State // contract state
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		data:          data,
		ip:            0,
		stack:         NewStack(128),
		contractState: contractState,
	}
}

func (vm *VM) Run() error {
	for {
		instr := vm.data[vm.ip]
		if err := vm.Exec(Instruction(instr)); err != nil {
			fmt.Println(err)
		}
		vm.ip++
		if vm.ip >= len(vm.data) {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	fmt.Println("executing instruction:", instr)
	switch instr {
	case InstructionPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstructionPushByte:
		vm.stack.Push(vm.data[vm.ip-1])
	case InstructionAdd:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		c := a.(int) + b.(int)
		vm.stack.Push(c)
	case InstructionSub:
		b := vm.stack.Pop()
		a := vm.stack.Pop()
		c := a.(int) - b.(int)
		vm.stack.Push(c)
	case InstructionPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[n-i-1] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
	case InstructionStore:
		v := vm.stack.Pop()
		var value []byte
		switch v.(type) {
		case int:
			value = util.SerializeInt64(int64(v.(int)))
		case []byte:
			value = v.([]byte)
		case byte:
			value = []byte{v.(byte)}
		}
		key := vm.stack.Pop().([]byte)
		vm.contractState.Put(key, value)
		fmt.Printf("key: %v, value: %v\n", key, value)
	}

	return nil
}

type Instruction byte

const (
	InstructionPushInt  Instruction = 0x0a // 10
	InstructionAdd      Instruction = 0x0b // 11
	InstructionPushByte Instruction = 0x0c // 12
	InstructionPack     Instruction = 0x0d // 13
	InstructionSub      Instruction = 0x0e // 14
	InstructionStore    Instruction = 0x0f // 15
)
