package structure

import (
	"cmp"
	"slices"
	"sort"

	"golang.org/x/exp/maps"
)

type WriteEfficientSortedMap[K cmp.Ordered, T cmp.Ordered] struct {
	store           map[K]T
	reversed        bool
	cacheSortedKeys []K
	cacheFresh      bool
}

func (s *WriteEfficientSortedMap[K, T]) Less(i int, j int) bool {
	return s.store[s.cacheSortedKeys[i]] < s.store[s.cacheSortedKeys[j]]
}

func (s *WriteEfficientSortedMap[K, T]) Swap(i int, j int) {
	key1 := s.cacheSortedKeys[i]
	key2 := s.cacheSortedKeys[j]

	s.cacheSortedKeys[j] = key1
	s.cacheSortedKeys[i] = key2
}

func NewWriteEfficientSortedMap[K cmp.Ordered, T cmp.Ordered](reversed bool) *WriteEfficientSortedMap[K, T] {
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
	s.cacheSortedKeys = maps.Keys(s.store)

	sort.Sort(s)

	s.cacheFresh = true
}

func (s *WriteEfficientSortedMap[K, T]) Get(index int) (key K, value T) {
	if s.reversed {
		index = len(s.store) - 1 - index
	}

	if !s.cacheFresh {
		s.Cache()
	}

	key = s.cacheSortedKeys[index]

	return key, s.store[key]
}

func (s *WriteEfficientSortedMap[K, T]) GetPosition(key K) int {
	if !s.cacheFresh {
		s.Cache()
	}

	index := slices.Index(s.cacheSortedKeys, key)
	if s.reversed {
		index = len(s.cacheSortedKeys) - index
	}

	return index
}
