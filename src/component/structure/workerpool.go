package structure

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type WorkerPool[T any] struct {
	queue              chan T
	workers            int
	workerFunction     func(T)
	stop               chan struct{}
	stopWG             sync.WaitGroup
	counter            int
	lastCounter        int
	lastPrintedCounter time.Time
	profilerOptions    ProfilerOptions
}

func NewWorkerPool[T any](workers int, workerFunction func(T), profilerOptions ProfilerOptions) *WorkerPool[T] {
	workerPool := WorkerPool[T]{
		queue:              make(chan T),
		workers:            workers,
		workerFunction:     workerFunction,
		stop:               make(chan struct{}),
		stopWG:             sync.WaitGroup{},
		counter:            0,
		lastCounter:        0,
		lastPrintedCounter: time.Now(),
		profilerOptions:    profilerOptions,
	}

	go workerPool.startWorkerThreads()

	return &workerPool
}

func (c *WorkerPool[T]) AddData(data T) {
	c.queue <- data
}

func (c *WorkerPool[T]) workerThread() {
	shouldStop := false
	alertsSeries := PerformanceManager.GetSeries("alerts")
	memory := c.profilerOptions.Has(MemoryProfilerOptionFlag)
	alerts := c.profilerOptions.Has(AlertsProfilerOptionFlag)

loop:
	for {
		if shouldStop && len(c.queue) == 0 {
			break loop
		}

		select {
		case data := <-c.queue:
			c.workerFunction(data)
			c.counter += 1

			profileData := []int{}

			if alerts {
				profileData = append(profileData, c.counter)
			}

			if memory {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				profileData = append(profileData, int(m.HeapInuse)+int(m.StackInuse))
			}

			if alerts || memory {
				alertsSeries.Add(time.Now(), profileData...)
			}

		case <-c.stop:
			shouldStop = true
		}
	}
}

func (c *WorkerPool[T]) startWorkerThreads() {
	c.stopWG.Add(c.workers)

	for range c.workers {
		go func() {
			c.workerThread()
			c.stopWG.Done()
		}()
	}

	go func() {
		for {
			now := time.Now()
			alertsPerSecond := (float32(c.counter) - float32(c.lastCounter)) / float32(now.Sub(c.lastPrintedCounter).Seconds())
			fmt.Printf("\r%d alerts (%.02f alerts/s) using %d workers    ", c.counter, alertsPerSecond, c.workers)
			c.lastCounter = c.counter
			c.lastPrintedCounter = now

			if c.workers == 0 {
				break
			}

			time.Sleep(time.Millisecond * 500)
		}
	}()
}

func (c *WorkerPool[T]) Stop() {
	for range c.workers {
		c.stop <- struct{}{}
	}
	c.stopWG.Wait()
	c.workers = 0
}
