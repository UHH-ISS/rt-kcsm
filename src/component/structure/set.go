package structure

import (
	"fmt"
	"strings"
)

type SetElement comparable

type Set[T SetElement] map[T]struct{}

func NewSet[T SetElement](elements ...T) Set[T] {
	s := Set[T]{}

	s.Append(elements...)

	return s
}

func (set Set[T]) Append(elements ...T) {
	for _, element := range elements {
		set[element] = struct{}{}
	}
}

func (set Set[T]) First() *T {
	for key := range set {
		return &key
	}

	return nil
}

func (set Set[T]) ToSlice() []T {
	elements := []T{}

	for key := range set {
		elements = append(elements, key)
	}

	return elements
}

func (set Set[T]) String() string {
	elements := []string{}

	for key := range set {
		elements = append(elements, fmt.Sprintf("%v", key))
	}

	return fmt.Sprintf("{%s}", strings.Join(elements, ", "))
}

func (set Set[T]) Intersect(otherSet Set[T]) Set[T] {
	elements := []T{}

	for key := range set {
		if _, ok := otherSet[key]; ok {
			elements = append(elements, key)
		}
	}

	return NewSet[T](elements...)
}

func (set Set[T]) Equal(otherSet Set[T]) bool {
	if otherSet.Size() != set.Size() {
		return false
	}

	for key := range set {
		if _, ok := otherSet[key]; !ok {
			return false
		}
	}

	return true
}

func (set Set[T]) Union(otherSets ...Set[T]) {
	for _, otherSet := range otherSets {
		for key := range otherSet {
			set[key] = struct{}{}
		}
	}
}

func (set Set[T]) Size() int {
	return len(set)
}
