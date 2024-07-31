package structure

import (
	"cmp"
	"slices"
)

type WriteEfficientSortedMap[K comparable, T cmp.Ordered] struct {
	store             map[K]T
	reversed          bool
	cacheSortedKeys   []K
	cacheSortedValues []T
	cacheFresh        bool
}

func NewWriteEfficientSortedMap[K comparable, T cmp.Ordered](reversed bool) *WriteEfficientSortedMap[K, T] {
	return &WriteEfficientSortedMap[K, T]{
		store:      map[K]T{},
		reversed:   reversed,
		cacheFresh: false,
	}
}

func (s *WriteEfficientSortedMap[K, T]) Len() int {
	return len(s.store)
}

func (s *WriteEfficientSortedMap[K, T]) Delete(key K) {
	delete(s.store, key)
	s.cacheFresh = false
}

func (s *WriteEfficientSortedMap[K, T]) Insert(key K, value T) {
	s.store[key] = value
	s.cacheFresh = false
}

func (s *WriteEfficientSortedMap[K, T]) Cache() {
	keys := []K{}
	values := []T{}

	for key, value := range s.store {
		insertPosition, _ := slices.BinarySearch(values, value)

		keys = slices.Insert(keys, insertPosition, key)
		values = slices.Insert(values, insertPosition, value)
	}

	s.cacheFresh = true
	s.cacheSortedKeys = keys
	s.cacheSortedValues = values
}

func (s *WriteEfficientSortedMap[K, T]) Get(index int) (key K, value T) {
	if s.reversed {
		index = len(s.store) - 1 - index
	}

	if !s.cacheFresh {
		s.Cache()
	}

	return s.cacheSortedKeys[index], s.cacheSortedValues[index]
}

func (s *WriteEfficientSortedMap[K, T]) GetPosition(key K) int {
	if !s.cacheFresh {
		s.Cache()
	}

	return slices.Index(s.cacheSortedKeys, key)
}
