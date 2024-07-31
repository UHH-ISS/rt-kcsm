package structure

import (
	"fmt"
	"strings"
)

type ProfilerOptions struct {
	flags   ProfilerOptionFlag
	graphID GraphID
}

func (p *ProfilerOptions) HasAny() bool {
	return p.flags != NoProfilerOptionFlag
}

func (p *ProfilerOptions) String() string {
	stringOptions := []string{}

	if p.Has(MemoryProfilerOptionFlag) {
		stringOptions = append(stringOptions, "memory")
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

	if p.Has(DoNotOptimizeGraphsProfilerOptionFlag) {
		stringOptions = append(stringOptions, "no-graph-optimization")
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
	return ProfilerOptions{}
}

func ParseProfilerOptions(stringOptions []string, profilerGraphID GraphID) ProfilerOptions {
	profilerOptions := ProfilerOptions{
		flags:   NoProfilerOptionFlag,
		graphID: profilerGraphID,
	}

	for _, stringOption := range stringOptions {
		if stringOption == "memory" {
			profilerOptions.flags |= MemoryProfilerOptionFlag
		} else if stringOption == "cpu" {
			profilerOptions.flags |= CpuProfilerOptionFlag
		} else if stringOption == "alerts" {
			profilerOptions.flags |= AlertsProfilerOptionFlag
		} else if stringOption == "graph-ranking" {
			profilerOptions.flags |= GraphRankingProfilerOptionFlag
		} else if stringOption == "graphs" {
			profilerOptions.flags |= GraphNumberProfilerOptionFlag
		} else if stringOption == "no-graph-optimization" {
			profilerOptions.flags |= DoNotOptimizeGraphsProfilerOptionFlag
		}
	}

	return profilerOptions
}

type ProfilerOptionFlag byte

const NoProfilerOptionFlag ProfilerOptionFlag = 0b0000_0000

const GraphRankingProfilerOptionFlag ProfilerOptionFlag = 0b0001_0000
const GraphNumberProfilerOptionFlag ProfilerOptionFlag = 0b0010_0000
const DoNotOptimizeGraphsProfilerOptionFlag ProfilerOptionFlag = 0b0100_0000

const MemoryProfilerOptionFlag ProfilerOptionFlag = 0b0000_0001
const CpuProfilerOptionFlag ProfilerOptionFlag = 0b0000_0010
const AlertsProfilerOptionFlag ProfilerOptionFlag = 0b0000_0100
