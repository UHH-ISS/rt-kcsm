package structure

import (
	"cmp"
)

type SortedMap[K comparable, T cmp.Ordered] interface {
	Insert(key K, value T)
	Delete(key K)
	Len() int
	Get(index int) (key K, value T)
	GetPosition(key K) int
}
