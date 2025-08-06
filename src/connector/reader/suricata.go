package reader

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"time"
)

type SuricataAlertReader[T structure.Stage, K structure.Stage] struct{}

type suricataLogEntry struct {
	Timestamp   string        `json:"timestamp"`
	Source      string        `json:"src_ip"`
	Destination string        `json:"dest_ip"`
	Alert       suricataAlert `json:"alert"`
	Label       string        `json:"label"`
}

type suricataAlert struct {
	Severity    int              `json:"severity"`
	Signature   string           `json:"signature"`
	SignatureId uint32           `json:"signature_id"`
	Metadata    suricataMetadata `json:"metadata"`
}

type suricataMetadata struct {
	Confidence []string `json:"confidence"`
}

const maxSeverityLevel = 4

var confidenceLevelMapping = map[string]float32{
	"Low":    0.25,
	"Medium": 0.5,
	"High":   1.0,
}

func (SR *SuricataAlertReader[T, K]) ChannelAlerts(rtkcsm behaviour.RTKCSM[T, K], reader io.ReadCloser) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var logEntry suricataLogEntry
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &logEntry); err != nil {
			log.Printf("error decoding file: %s\n", err)
		}

		if logEntry.Alert.Severity > 0 {
			timestamp, err := time.Parse("2006-01-02T15:04:05.000000-0700", logEntry.Timestamp)
			if err != nil {
				log.Printf("error parsing time: %s", err)
			}
			sourceIP := structure.ParseIPAddress(logEntry.Source)
			destinationIP := structure.ParseIPAddress(logEntry.Destination)
			severity := float32((maxSeverityLevel)-logEntry.Alert.Severity) / (maxSeverityLevel - 1)

			// By default we use a confidence score of 1
			confidence := float32(1)
			if len(logEntry.Alert.Metadata.Confidence) > 0 {
				confidenceLevel := logEntry.Alert.Metadata.Confidence[0]

				var ok bool
				confidence, ok = confidenceLevelMapping[confidenceLevel]
				if !ok {
					log.Printf("error mapping confidence level: %s\n", confidenceLevel)
				}
			}

			alert := structure.Alert{
				Timestamp:     timestamp,
				SourceIP:      sourceIP,
				DestinationIP: destinationIP,
				Severity:      severity,
				Confidence:    confidence,
				SignatureId:   logEntry.Alert.SignatureId,
				Cause:         logEntry.Alert.Signature,
				Label:         logEntry.Label,
			}

			err = rtkcsm.AddAlert(alert)
			if err != nil {
				//log.Println(err)
			}
		}
	}

	return nil
}
