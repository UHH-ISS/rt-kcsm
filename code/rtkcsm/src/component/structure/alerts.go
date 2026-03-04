package structure

import (
	"time"
)

type Alert struct {
	Timestamp     time.Time `json:"timestamp"`
	SourceIP      IPAddress `json:"source_ip"`
	DestinationIP IPAddress `json:"destination_ip"`
	Severity      float32   `json:"severity"`
	Confidence    float32   `json:"confidence"`
	Cause         string    `json:"cause"`
	SignatureId   uint32    `json:"signature_id"`
	Label         string    `json:"label"`
}

type Alerts []Alert

func (a Alerts) Len() int           { return len(a) }
func (a Alerts) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Alerts) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }
