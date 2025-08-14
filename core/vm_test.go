package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVM_Run(t *testing.T) {
	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	// 3
	// push stack
	data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}
	vm := NewVM(data)
	err := vm.Run()
	require.NoError(t, err)
	assert.Equal(t, 2, vm.sp)
	assert.Equal(t, byte(3), vm.stack[vm.sp])
}
