package core

import "fmt"

type VM struct {
	data  []byte
	ip    int // instruction pointer
	stack []byte
	sp    int // stack pointer
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: make([]byte, 1024), // TODO
		sp:    -1,
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
	case InstructionPush:
		v.pushStack(v.data[v.ip-1])
	case InstructionAdd:
		a := v.stack[v.sp-1]
		b := v.stack[v.sp]
		c := a + b
		v.pushStack(c)
	}

	return nil
}

func (v *VM) pushStack(b byte) {
	v.sp++
	v.stack[v.sp] = b
}

type Instruction byte

const (
	InstructionPush Instruction = 0x0a // 10
	InstructionAdd  Instruction = 0x0b // 11
)
