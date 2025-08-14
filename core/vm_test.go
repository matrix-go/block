package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStack_Shift(t *testing.T) {
	s := NewStack(128)
	s.Push(1)
	s.Push(2)
	s.Push(3)

	t.Logf("stack: %v", s.data)
	v := s.Shift()
	assert.Equal(t, 1, v)
	t.Logf("stack: %v", s.data)
}

func TestVM_Run(t *testing.T) {
	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	// 3
	// push stack int and add
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	vm := NewVM(data)
	err := vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)
	result := vm.stack.Shift()
	assert.Equal(t, 3, result)

	// push bytes and pack
	data = []byte{0x02, 0x0a, 0x61, 0x0c, 0x61, 0x0c, 0x0d}
	vm = NewVM(data)
	err = vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)
	res := vm.stack.Shift().([]byte)
	assert.Equal(t, "aa", string(res))

	// push int and sub
	data = []byte{0x02, 0x0a, 0x01, 0x0a, 0x0e}
	vm = NewVM(data)
	err = vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)
	r := vm.stack.Shift().(int)
	assert.Equal(t, 1, r)
}
