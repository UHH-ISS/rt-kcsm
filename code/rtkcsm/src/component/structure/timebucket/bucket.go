package timebucket

import (
	"time"
)

type Bucket[T comparable] struct {
	store map[T]time.Time
}

func NewBucket[T comparable]() *Bucket[T] {
	return &Bucket[T]{
		store: make(map[T]time.Time, BUCKET_SIZE),
	}
}

func (t *Bucket[T]) Set(value T, time time.Time) {
	t.store[value] = time
}

func (t *Bucket[T]) Get(value T) time.Time {
	return t.store[value]
}

func (t *Bucket[T]) Delete(value T) {
	delete(t.store, value)
}

func (t *Bucket[T]) Len() int {
	return len(t.store)
}

func (t *Bucket[T]) GetAllBefore(time time.Time) []T {
	keys := []T{}
	for key, storedTime := range t.store {
		if !storedTime.After(time) {
			keys = append(keys, key)
		}
	}

	return keys
}
