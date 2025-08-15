package core

import (
	"github.com/matrix-go/block/util"
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

	contractState := NewState()
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	// 1 + 2
	vm := NewVM(data, contractState)
	err := vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)

	result := util.DeserializeInt64(vm.stack.Shift().([]byte))
	assert.Equal(t, int64(3), result)

	// push bytes and pack
	// 0x61(a), 0x0c(pushByte), 0x61(a), 0x0c(pushByte), 0x02(len=2), 0x0a(pushInt), 0x0d(pack)
	// aa
	data = []byte{0x61, 0x0c, 0x61, 0x0c, 0x02, 0x0a, 0x0d}
	vm = NewVM(data, contractState)
	err = vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)
	res := vm.stack.Shift().([]byte)
	assert.Equal(t, "aa", string(res))

	// push int and sub
	// 2-1
	data = []byte{0x02, 0x0a, 0x01, 0x0a, 0x0e}
	vm = NewVM(data, contractState)
	err = vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 1, vm.stack.sp)
	t.Logf("stack: %v", vm.stack.data)
	r := util.DeserializeInt64(vm.stack.Shift().([]byte))
	assert.Equal(t, int64(1), r)

	// store state
	// push FOO and pack
	// push 3 push 2 and sub
	// store [FOO, 1]
	// 0x03(len=3), 0x0a(pushInt), 0x46(f), 0x0c(pushByte), 0x4f(o), 0x0c(pushByte), 0x4f(o), 0x0c(pushByte), 0x0d(pack)
	// foo = 3-2
	//data = []byte{
	//	0x03, 0x0a, 0x02, 0x0a, 0x0e, // push 3, push 2 and sub
	//	0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, // push FOO and pack
	//	0x0f, // store [FOO,1]
	//}
	//
	//vm = NewVM(data, contractState)
	//err = vm.Run()
	//require.NoError(t, err)
	//t.Logf("stack: %v", vm.stack.data)
	//t.Logf("stack sp: %v", vm.stack.sp)
	//assert.Equal(t, 1, r)
	//t.Logf("state: %v", vm.contractState.data)
	//val, err := vm.contractState.Get([]byte("FOO"))
	//require.NoError(t, err)
	//des := util.DeserializeInt64(val)
	//assert.Equal(t, int64(1), des)

	data = []byte{
		0x03, 0x0a, 0x02, 0x0a, 0x0e, // push 3, push 2 and sub
		0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, // push FOO and pack
		0x0f,                         // store [FOO,1]
		0x03, 0x0a, 0x02, 0x0a, 0x0b, // push 3, push 2 and add
		0x46, 0x0c, 0x4f, 0x0c, 0x4d, 0x0c, 0x03, 0x0a, 0x0d, // push FOM and pack
		0x0f, // store [FOM,1]
	}

	vm = NewVM(data, contractState)
	err = vm.Run()
	require.NoError(t, err)
	t.Logf("stack: %v", vm.stack.data)
	t.Logf("stack sp: %v", vm.stack.sp)
	assert.Equal(t, int64(1), r)
	t.Logf("state: %v", vm.contractState.data)
	val, err := vm.contractState.Get([]byte("FOM"))
	require.NoError(t, err)
	des := util.DeserializeInt64(val)
	assert.Equal(t, int64(5), des)

	data = []byte{
		0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, // push FOO and pack
		0x10, // get FOO
	}

	vm = NewVM(data, contractState)
	err = vm.Run()
	require.NoError(t, err)
	t.Logf("stack: %v", vm.stack.data)
	t.Logf("stack sp: %v", vm.stack.sp)
	t.Logf("state: %v", vm.contractState.data)
	re := util.DeserializeInt64(vm.stack.Shift().([]byte))
	assert.Equal(t, int64(1), re)

	//
	data = []byte{
		0x02, 0x0a, 0x03, 0x0a, 0x11, // 2 * 3
		0x46, 0x0c, 0x4f, 0x0c, 0x02, 0x0a, 0x0d, // push FO and pack
		0x0f,                                     // store [FO, 6]
		0x46, 0x0c, 0x4f, 0x0c, 0x02, 0x0a, 0x0d, // push FOO and pack
		0x10, // get FO
	}

	vm = NewVM(data, contractState)
	err = vm.Run()
	require.NoError(t, err)
	t.Logf("stack: %v", vm.stack.data)
	t.Logf("stack sp: %v", vm.stack.sp)
	t.Logf("state: %v", vm.contractState.data)
	re = util.DeserializeInt64(vm.stack.Shift().([]byte))
	assert.Equal(t, int64(6), re)

}
