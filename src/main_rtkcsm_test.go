package main

import (
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/visualization"
	"sort"
	"testing"
	"time"
)

var profilerOptions = structure.NewProfilerOptions()

func TestGraphGeneration(t *testing.T) {
	rtkcsm := behaviour.NewIncrementalRTKCSM(1, structure.NewSimplifiedUKCStageMapper(), structure.NewUKCStateMachine[structure.SimplifiedUKCStage](), &profilerOptions)

	seconds := time.Now().Unix()

	alerts := structure.Alerts{
		// Malformed alert
		structure.Alert{
			Timestamp:     time.Unix(seconds-1, 0),
			SourceIP:      structure.ParseIPAddress("0.0.0.0"),
			DestinationIP: structure.ParseIPAddress("172.31.64.67"),
			Severity:      1,
			Confidence:    1,
		},
		// False positive alert
		structure.Alert{
			Timestamp:     time.Unix(seconds-1, 0),
			SourceIP:      structure.ParseIPAddress("1.1.13.37"),
			DestinationIP: structure.ParseIPAddress("172.31.64.67"),
			Severity:      1,
			Confidence:    1,
		},
		// Correct alerts
		structure.Alert{
			Timestamp:     time.Unix(seconds-5, 0),
			SourceIP:      structure.ParseIPAddress("1.1.13.37"),
			DestinationIP: structure.ParseIPAddress("172.31.64.67"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("12.34.12.34"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-3, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("1.1.14.47"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-1, 0),
			SourceIP:      structure.ParseIPAddress("172.31.69.20"),
			DestinationIP: structure.ParseIPAddress("1.1.15.57"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("172.31.69.20"),
			Severity:      0.5,
			Confidence:    1,
		},
		// Same source and destination with higher severity
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("172.31.69.20"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("172.31.69.20"),
			Severity:      0.5,
			Confidence:    1,
		},
		// Duplicate alert
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.31.64.67"),
			DestinationIP: structure.ParseIPAddress("172.31.69.20"),
			Severity:      0.5,
			Confidence:    1,
		},
	}

	sort.Sort(alerts)

	for _, alert := range alerts {
		rtkcsm.AddAlert(alert)
	}

	sortedGraphList := rtkcsm.GetGraphList(-1)

	if sortedGraphList.Count != 2 {
		t.Errorf("graph list too short or too long: %d graphs", sortedGraphList.Count)
	} else {
		graph := rtkcsm.GetGraph(sortedGraphList.Graphs[0].ID)
		if graph.Relevance() != 0.75 {
			t.Errorf("graph does not have all stages: relevance=%f", graph.Relevance())
		}
	}
}

func TestGraphGenerationSpecialGraph(t *testing.T) {
	rtkcsm := behaviour.NewIncrementalRTKCSM(1, structure.NewSimplifiedUKCStageMapper(), structure.NewUKCStateMachine[structure.SimplifiedUKCStage](), &profilerOptions)

	seconds := time.Now().Unix()

	alerts := structure.Alerts{
		structure.Alert{
			Timestamp:     time.Unix(seconds, 0),
			SourceIP:      structure.ParseIPAddress("192.168.10.8"),
			DestinationIP: structure.ParseIPAddress("192.168.10.50"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-1, 0),
			SourceIP:      structure.ParseIPAddress("192.168.10.8"),
			DestinationIP: structure.ParseIPAddress("205.174.165.73"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-2, 0),
			SourceIP:      structure.ParseIPAddress("205.174.165.73"),
			DestinationIP: structure.ParseIPAddress("192.168.10.8"),
			Severity:      1,
			Confidence:    1,
		},
	}

	sort.Sort(alerts)

	for _, alert := range alerts {
		rtkcsm.AddAlert(alert)
	}

	sortedGraphList := rtkcsm.GetGraphList(-1)

	if sortedGraphList.Count != 1 {
		t.Errorf("graph list too short or too long: %d graphs", sortedGraphList.Count)
	}
}

func TestGraphGenerationComplexAttack(t *testing.T) {
	rtkcsm := behaviour.NewIncrementalRTKCSM(1, structure.NewSimplifiedUKCStageMapper(), structure.NewUKCStateMachine[structure.SimplifiedUKCStage](), &profilerOptions)

	seconds := time.Now().Unix()

	alerts := structure.Alerts{
		structure.Alert{
			Timestamp:     time.Unix(seconds-5, 0),
			SourceIP:      structure.ParseIPAddress("94.141.120.36"),
			DestinationIP: structure.ParseIPAddress("172.16.42.42"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-4, 0),
			SourceIP:      structure.ParseIPAddress("172.16.42.42"),
			DestinationIP: structure.ParseIPAddress("218.92.0.27"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-3, 0),
			SourceIP:      structure.ParseIPAddress("172.16.42.42"),
			DestinationIP: structure.ParseIPAddress("172.16.42.1"),
			Severity:      1,
			Confidence:    1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds-2, 0),
			SourceIP:      structure.ParseIPAddress("172.16.42.1"),
			DestinationIP: structure.ParseIPAddress("10.12.2.93"),
			Severity:      1,
			Confidence:    1,
		},
	}

	sort.Sort(alerts)

	for _, alert := range alerts {
		rtkcsm.AddAlert(alert)
	}

	sortedGraphList := rtkcsm.GetGraphList(-1)

	if sortedGraphList.Count != 1 {
		t.Errorf("graph list too short or too long: %d graphs", sortedGraphList.Count)
		visualization.Start(":8080", rtkcsm, assets)
	}
}
