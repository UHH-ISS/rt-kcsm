package structure

import (
	"fmt"
	"io"
	"log"
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
	entries Measurements[T]
	channel chan Measurement[T]
	columns []string
	wait    *sync.WaitGroup
}

func newMeasurementSeries[T any](columns []string) MeasurementSeries[T] {
	series := MeasurementSeries[T]{
		entries: []Measurement[T]{},
		channel: make(chan Measurement[T]),
		columns: append([]string{"time"}, columns...),
		wait:    &sync.WaitGroup{},
	}

	return series
}

func (m *MeasurementSeries[T]) Start(writer io.Writer) error {
	_, err := writer.Write([]byte(strings.Join(m.columns, ",") + "\n"))
	if err != nil {
		return err
	}

	m.wait.Add(1)
	go func() {
		for {
			entry, ok := <-m.channel
			if !ok {
				break
			}
			array := []string{}

			for _, point := range entry.Data {
				array = append(array, fmt.Sprintf("%v", point))
			}

			_, err := writer.Write([]byte(fmt.Sprintf("%s,%s\n", entry.Time, strings.Join(array, ","))))
			if err != nil {
				log.Printf("error writing line: %s\n", err)
				break
			}
		}

		m.wait.Done()
	}()

	return nil
}

func (m *MeasurementSeries[T]) Stop() {
	close(m.channel)
	m.wait.Wait()
}

func (m *MeasurementSeries[T]) Add(time time.Time, data ...T) {
	m.channel <- Measurement[T]{
		Data: data,
		Time: time,
	}
}

type MeasurementManager[T any] struct {
	series map[string]*MeasurementSeries[T]
	mutex  sync.RWMutex
}

func (m *MeasurementManager[T]) AddSeries(name string, columns []string) *MeasurementSeries[T] {
	series := newMeasurementSeries[T](columns)

	m.mutex.Lock()
	m.series[name] = &series
	m.mutex.Unlock()

	return &series
}

func (m *MeasurementManager[T]) GetSeries(name string) *MeasurementSeries[T] {
	m.mutex.RLock()
	series := m.series[name]
	m.mutex.RUnlock()

	return series
}

func (m *MeasurementManager[T]) StopAllSeries() {
	m.mutex.RLock()
	for _, series := range m.series {
		series.Stop()
	}
	m.mutex.RUnlock()
}

func newMeasurementManager[T any]() MeasurementManager[T] {
	return MeasurementManager[T]{
		series: map[string]*MeasurementSeries[T]{},
		mutex:  sync.RWMutex{},
	}
}

var PerformanceManager = newMeasurementManager[int]()
