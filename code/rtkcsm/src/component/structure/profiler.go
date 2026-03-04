package structure

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
)

type ProfilerOptions struct {
	flags              ProfilerOptionFlag
	graphID            GraphID
	alertSeries        *MeasurementSeries[int]
	memorySeries       *MeasurementSeries[int]
	graphCountSeries   *MeasurementSeries[int]
	counter            int
	lastCounter        int
	lastPrintedCounter time.Time
	logResolution      int
}

func (p *ProfilerOptions) HasAny() bool {
	return p.flags != NoProfilerOptionFlag
}

func (p *ProfilerOptions) String() string {
	stringOptions := []string{}

	if p.Has(ProgressProfilerOptionFlag) {
		stringOptions = append(stringOptions, "progress")
	}

	if p.Has(MemoryProfilerOptionFlag) {
		stringOptions = append(stringOptions, "memory")
	}

	if p.Has(MemoryAllocationProfilerOptionFlag) {
		stringOptions = append(stringOptions, "memory-heap")
	}

	if p.Has(CpuProfilerOptionFlag) {
		stringOptions = append(stringOptions, "cpu")
	}

	if p.Has(AlertsProfilerOptionFlag) {
		stringOptions = append(stringOptions, "alerts")
	}

	if p.Has(GraphRankingProfilerOptionFlag) {
		stringOptions = append(stringOptions, fmt.Sprintf("graph-ranking (%v)", p.graphID))
	}

	if p.Has(GraphNumberProfilerOptionFlag) {
		stringOptions = append(stringOptions, "graphs")
	}

	return strings.Join(stringOptions, ", ")
}

func (p *ProfilerOptions) Has(otherP ProfilerOptionFlag) bool {
	return p.flags&otherP == otherP
}

func (p *ProfilerOptions) GetGraphId() GraphID {
	return p.graphID
}

func NewProfilerOptions() ProfilerOptions {
	return ProfilerOptions{
		logResolution: 1000,
	}
}

func ParseProfilerOptions(stringOptions []string, profilerGraphID GraphID, logResolution int) ProfilerOptions {
	profilerOptions := ProfilerOptions{
		flags:         NoProfilerOptionFlag,
		graphID:       profilerGraphID,
		logResolution: logResolution,
	}

	for _, stringOption := range stringOptions {
		profilerOptions.flags |= optionToFlag[ProfilerOption(stringOption)]
	}

	if profilerOptions.Has(MemoryProfilerOptionFlag) {
		profilerOptions.memorySeries = PerformanceManager.AddSeries("memory", []string{"alerts", "memory"})
	}

	if profilerOptions.Has(AlertsProfilerOptionFlag) {
		profilerOptions.alertSeries = PerformanceManager.AddSeries("alerts", []string{"alerts"})
	}

	if profilerOptions.Has(GraphNumberProfilerOptionFlag) {
		profilerOptions.graphCountSeries = PerformanceManager.AddSeries("graph-count", []string{"count"})
	}

	if profilerOptions.Has(GraphRankingProfilerOptionFlag) {
		PerformanceManager.AddSeries("graph-ranking", []string{"rank", "count"})
	}

	return profilerOptions
}

func (p *ProfilerOptions) TakeMeasurement(graphCount int, final bool) error {
	if p.counter%p.logResolution == 0 || final {
		now := time.Now()

		if p.memorySeries != nil {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			total := m.HeapInuse + m.StackInuse
			if total > math.MaxInt {
				return fmt.Errorf("total memory usage exceeds int: %d < %d", total, math.MaxInt32)
			}
			p.memorySeries.Add(now, p.counter, int(total))
		}

		if p.alertSeries != nil {
			p.alertSeries.Add(now, p.counter)
		}

		if p.graphCountSeries != nil {
			p.graphCountSeries.Add(now, graphCount)
		}
	}

	if p.Has(ProgressProfilerOptionFlag) && (time.Since(p.lastPrintedCounter) > time.Millisecond*500 || final) {
		now := time.Now()
		alertsPerSecond := (float32(p.counter) - float32(p.lastCounter)) / float32(now.Sub(p.lastPrintedCounter).Seconds())
		fmt.Printf("\r%d alerts (%.02f alerts/s)", p.counter, alertsPerSecond)
		p.lastPrintedCounter = now
		p.lastCounter = p.counter

		if final {
			fmt.Println()
		}
	}

	p.counter += 1

	return nil
}

type ProfilerOptionFlag byte

const NoProfilerOptionFlag ProfilerOptionFlag = 0b0000_0000

const MemoryProfilerOptionFlag ProfilerOptionFlag = 0b0000_0001
const MemoryAllocationProfilerOptionFlag ProfilerOptionFlag = 0b0000_0010
const CpuProfilerOptionFlag ProfilerOptionFlag = 0b0000_0100
const AlertsProfilerOptionFlag ProfilerOptionFlag = 0b0000_1000

const GraphRankingProfilerOptionFlag ProfilerOptionFlag = 0b0001_0000
const GraphNumberProfilerOptionFlag ProfilerOptionFlag = 0b0010_0000

const ProgressProfilerOptionFlag ProfilerOptionFlag = 0b1000_0000

type ProfilerOption string

const MemoryOption ProfilerOption = "memory"
const MemoryAllocationOption ProfilerOption = "memory-alloc"
const CpuOption ProfilerOption = "cpu"
const AlertsOption ProfilerOption = "alerts"

const GraphRankingOption ProfilerOption = "graph-ranking"
const GraphNumberOption ProfilerOption = "graphs"

const ProgressOption ProfilerOption = "progress"

var optionToFlag = map[ProfilerOption]ProfilerOptionFlag{
	MemoryOption:           MemoryProfilerOptionFlag,
	MemoryAllocationOption: MemoryAllocationProfilerOptionFlag,
	CpuOption:              CpuProfilerOptionFlag,
	AlertsOption:           AlertsProfilerOptionFlag,
	GraphRankingOption:     GraphRankingProfilerOptionFlag,
	GraphNumberOption:      GraphNumberProfilerOptionFlag,
	ProgressOption:         ProgressProfilerOptionFlag,
}
