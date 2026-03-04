package timebucket

import (
	"sync"
	"time"
)

type LeafType string

const LeftLeafType LeafType = "left"
const RightLeafType LeafType = "right"
const NoneLeafType LeafType = "none"

const BUCKET_SIZE = 1000

type TimeBucketIndex[T comparable] struct {
	index        []time.Time
	buckets      []*Bucket[T]
	bucketsMutex *sync.RWMutex
}

func NewTimeBucketIndex[T comparable]() *TimeBucketIndex[T] {
	t := &TimeBucketIndex[T]{
		index:        []time.Time{},
		buckets:      []*Bucket[T]{},
		bucketsMutex: &sync.RWMutex{},
	}

	return t
}

func (t TimeBucketIndex[T]) GetAllBefore(time time.Time) []T {
	if len(t.buckets) > 1 && t.buckets[0].Len() == 0 {
		t.deleteBucket(0)
	}

	values := []T{}
	index := len(t.index) - 1

	t.bucketsMutex.RLock()
	defer t.bucketsMutex.RUnlock()

	for {
		if index >= 0 {
			if !t.index[index].After(time) {
				bucketValues := t.buckets[index].GetAllBefore(time)
				values = append(values, bucketValues...)
			}
			index -= 1
		} else {
			return values
		}
	}
}

func (t *TimeBucketIndex[T]) deleteBucket(index int) {
	t.bucketsMutex.Lock()
	defer t.bucketsMutex.Unlock()

	t.index = append(t.index[:index], t.index[index+1:]...)
	t.buckets = append(t.buckets[:index], t.buckets[index+1:]...)
}

func (t *TimeBucketIndex[T]) addBucket(value T, time time.Time) *Bucket[T] {
	t.bucketsMutex.Lock()
	defer t.bucketsMutex.Unlock()

	bucket := NewBucket[T]()
	bucket.Set(value, time)
	t.buckets = append(t.buckets, bucket)
	t.index = append(t.index, time)
	return bucket
}

func (t *TimeBucketIndex[T]) Insert(value T, time time.Time) *Bucket[T] {
	if len(t.buckets) == 0 {
		return t.addBucket(value, time)
	} else {
		index := len(t.buckets) - 1
		bucket := t.buckets[index]
		smallest := t.index[index]

		if bucket.Len() < BUCKET_SIZE {
			bucket.Set(value, time)
			if smallest.After(time) {
				t.index[index] = time
			}
			return bucket
		} else {
			return t.addBucket(value, time)
		}
	}
}
