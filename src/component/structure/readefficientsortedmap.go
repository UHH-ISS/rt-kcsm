package structure

import (
	"cmp"
	"slices"
)

type ReadEfficientSortedMap[K comparable, T cmp.Ordered] struct {
	keys     []K
	values   []T
	reversed bool
}

func NewReadEfficientSortedMap[K comparable, T cmp.Ordered](reversed bool) *ReadEfficientSortedMap[K, T] {
	return &ReadEfficientSortedMap[K, T]{
		keys:     []K{},
		values:   []T{},
		reversed: reversed,
	}
}

func (s *ReadEfficientSortedMap[K, T]) Len() int {
	return len(s.keys)
}

func (s *ReadEfficientSortedMap[K, T]) Get(index int) (K, T) {
	if s.reversed {
		index = len(s.keys) - 1 - index
	}
	return s.keys[index], s.values[index]
}

func (s *ReadEfficientSortedMap[K, T]) GetPosition(key K) int {
	index := slices.Index(s.keys, key)
	if s.reversed {
		index = len(s.keys) - index
	}

	return index
}

func (s *ReadEfficientSortedMap[K, T]) Delete(key K) {
	index := slices.Index(s.keys, key)

	if index >= 0 {
		s.keys = slices.Delete(s.keys, index, index+1)
		s.values = slices.Delete(s.values, index, index+1)
	}
}

func (s *ReadEfficientSortedMap[K, T]) Insert(key K, value T) {
	s.Delete(key)

	index, _ := slices.BinarySearch(s.values, value)

	s.keys = slices.Insert(s.keys, index, key)
	s.values = slices.Insert(s.values, index, value)
}
