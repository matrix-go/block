package core

import "fmt"

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
	ret := s.data[s.sp]
	s.data = append(s.data[:s.sp], s.data[s.sp+1:]...)
	s.sp--
	return ret
}

func (s *Stack) Shift() any {
	ret := s.data[0]
	s.data = append(s.data[1:], nil)
	s.sp--
	return ret
}

type VM struct {
	data  []byte
	ip    int // instruction pointer
	stack *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: NewStack(128),
	}
}

func (v *VM) Run() error {
	for {
		instr := v.data[v.ip]
		if err := v.Exec(Instruction(instr)); err != nil {
			fmt.Println(err)
		}
		v.ip++
		if v.ip >= len(v.data) {
			break
		}
	}
	return nil
}

func (v *VM) Exec(instr Instruction) error {
	fmt.Println("executing instruction:", instr)
	switch instr {
	case InstructionPushInt:
		v.stack.Push(int(v.data[v.ip-1]))
	case InstructionPushByte:
		v.stack.Push(v.data[v.ip-1])
	case InstructionAdd:
		a := v.stack.Shift()
		b := v.stack.Shift()
		c := a.(int) + b.(int)
		v.stack.Push(c)
	case InstructionSub:
		a := v.stack.Shift()
		b := v.stack.Shift()
		c := a.(int) - b.(int)
		v.stack.Push(c)
	case InstructionPack:
		n := v.stack.Shift().(int)
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[i] = v.stack.Shift().(byte)
		}
		v.stack.Push(b)
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
)
