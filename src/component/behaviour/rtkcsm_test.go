package behaviour

import (
	"rtkcsm/component/structure"
	"sort"
	"testing"
	"time"
)

func TestGraphGeneration(t *testing.T) {
	for i := range 12 {
		rtkcsm := NewIncrementalRTKCSM(1, structure.NewProfilerOptions())

		seconds := time.Now().Unix()

		alerts := structure.Alerts{
			// Malformed alert
			structure.Alert{
				Timestamp:     time.Unix(seconds-1, 0),
				SourceIP:      structure.ParseIPAddress("0.0.0.0"),
				DestinationIP: structure.ParseIPAddress("172.31.64.67"),
				Severity:      1,
			},
			// False positive alert
			structure.Alert{
				Timestamp:     time.Unix(seconds-1, 0),
				SourceIP:      structure.ParseIPAddress("1.1.13.37"),
				DestinationIP: structure.ParseIPAddress("172.31.64.67"),
				Severity:      1,
			},
			// Correct alerts
			structure.Alert{
				Timestamp:     time.Unix(seconds-5, 0),
				SourceIP:      structure.ParseIPAddress("1.1.13.37"),
				DestinationIP: structure.ParseIPAddress("172.31.64.67"),
				Severity:      1,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-4, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("12.34.12.34"),
				Severity:      1,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-3, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("1.1.14.47"),
				Severity:      1,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-1, 0),
				SourceIP:      structure.ParseIPAddress("172.31.69.20"),
				DestinationIP: structure.ParseIPAddress("1.1.15.57"),
				Severity:      1,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-4, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("172.31.69.20"),
				Severity:      0.5,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-4, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("172.31.69.20"),
				Severity:      1,
			},
			structure.Alert{
				Timestamp:     time.Unix(seconds-4, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("172.31.69.20"),
				Severity:      0.5,
			},
			// Duplicate alert
			structure.Alert{
				Timestamp:     time.Unix(seconds-4, 0),
				SourceIP:      structure.ParseIPAddress("172.31.64.67"),
				DestinationIP: structure.ParseIPAddress("172.31.69.20"),
				Severity:      0.5,
			},
		}

		sort.Sort(alerts)

		for _, alert := range alerts {
			rtkcsm.Add(alert)
		}

		rtkcsm.Stop([]structure.ExportFormat{})

		sortedGraphList := rtkcsm.GetGraphList(-1)

		if sortedGraphList.Count != 2 {
			t.Errorf("[workers=%d] graph list too short or too long: %d graphs", i, sortedGraphList.Count)
		} else {
			graph := rtkcsm.GetGraph(sortedGraphList.Graphs[0].ID)
			if graph.Relevance() != 4 {
				t.Errorf("[workers=%d] graph does not have all stages: relevance=%f", i, graph.Relevance())
			}
		}
	}
}

func TestGraphGenerationSpecialGraph(t *testing.T) {
	rtkcsm := NewIncrementalRTKCSM(1, structure.NewProfilerOptions())

	seconds := time.Now().Unix()

	alerts := structure.Alerts{
		structure.Alert{
			Timestamp:     time.Unix(seconds-1, 0),
			SourceIP:      structure.ParseIPAddress("192.168.10.8"),
			DestinationIP: structure.ParseIPAddress("205.174.165.73"),
			Severity:      1,
		},
		structure.Alert{
			Timestamp:     time.Unix(seconds, 0),
			SourceIP:      structure.ParseIPAddress("192.168.10.8"),
			DestinationIP: structure.ParseIPAddress("192.168.10.50"),
			Severity:      1,
		},
	}

	sort.Sort(alerts)

	for _, alert := range alerts {
		rtkcsm.Add(alert)
	}

	rtkcsm.Stop([]structure.ExportFormat{})

	sortedGraphList := rtkcsm.GetGraphList(-1)

	if sortedGraphList.Count != 1 {
		t.Errorf("graph list too short or too long: %d graphs", sortedGraphList.Count)
	}
}
