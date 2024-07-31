package structure

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

type Measurement[T any] struct {
	Data []T
	Time time.Time
}

type Measurements[T any] []Measurement[T]

func (m Measurements[T]) Len() int           { return len(m) }
func (a Measurements[T]) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Measurements[T]) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }

type MeasurementSeries[T any] struct {
	entries         Measurements[T]
	channel         chan Measurement[T]
	maxValueColumns int
}

func (m *MeasurementSeries[T]) Add(time time.Time, data ...T) {
	m.channel <- Measurement[T]{
		Data: data,
		Time: time,
	}
}

func (m *MeasurementSeries[T]) Save(writer io.Writer) {
	entries := m.entries
	sort.Sort(entries)

	header := []string{"time"}
	for i := range m.maxValueColumns {
		header = append(header, fmt.Sprintf("value_%d", i))
	}

	writer.Write([]byte(strings.Join(header, ",") + "\n"))

	for _, entry := range entries {
		array := []string{}

		for _, point := range entry.Data {
			array = append(array, fmt.Sprintf("%v", point))
		}

		writer.Write([]byte(fmt.Sprintf("%s,%s\n", entry.Time, strings.Join(array, ","))))
	}
}

type MeasurementManager[T any] struct {
	series map[string]*MeasurementSeries[T]
	mutex  sync.RWMutex
}

func (m *MeasurementManager[T]) addSeries(name string) *MeasurementSeries[T] {
	series := MeasurementSeries[T]{
		entries: []Measurement[T]{},
		channel: make(chan Measurement[T]),
	}

	go func() {
		for {
			entry := <-series.channel
			series.entries = append(series.entries, entry)
			if len(entry.Data) > series.maxValueColumns {
				series.maxValueColumns = len(entry.Data)
			}
		}
	}()

	m.mutex.Lock()
	m.series[name] = &series
	m.mutex.Unlock()

	return &series
}

func (m *MeasurementManager[T]) GetSeries(name string) *MeasurementSeries[T] {
	m.mutex.RLock()
	series, ok := m.series[name]
	m.mutex.RUnlock()

	if !ok {
		series = m.addSeries(name)
	}

	return series
}

func newMeasurementManager[T any]() MeasurementManager[T] {
	return MeasurementManager[T]{
		series: map[string]*MeasurementSeries[T]{},
		mutex:  sync.RWMutex{},
	}
}

var PerformanceManager = newMeasurementManager[int]()
