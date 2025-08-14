package types

import (
	"fmt"
	"reflect"
)

type List[T any] struct {
	Data []T
}

func NewList[T any]() *List[T] {
	return &List[T]{
		Data: make([]T, 0),
	}
}

func (l *List[T]) Get(idx int) T {
	if idx < 0 || idx > len(l.Data)-1 {
		err := fmt.Errorf("index(%d) out of range, (0,%d)", idx, len(l.Data)-1)
		panic(err)
	}
	return l.Data[idx]
}

func (l *List[T]) Insert(data T) {
	l.Data = append(l.Data, data)
}

func (l *List[T]) Clear() {
	l.Data = make([]T, 0)
}

func (l *List[T]) GetIndex(data T) int {
	for i, v := range l.Data {
		if reflect.DeepEqual(v, data) {
			return i
		}
	}
	return -1
}

func (l *List[T]) Remove(data T) {
	if idx := l.GetIndex(data); idx != -1 {
		l.Data = append(l.Data[:idx], l.Data[idx+1:]...)
	}
}

func (l *List[T]) Count() int {
	return len(l.Data)
}
