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

type SuricataTenzirAlertReader[T structure.Stage, K structure.Stage] struct{}

func (SR *SuricataTenzirAlertReader[T, K]) ChannelAlerts(rtkcsm behaviour.RTKCSM[T, K], reader io.ReadCloser) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var logEntry suricataLogEntry
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &logEntry); err != nil {
			log.Printf("error decoding file: %s\n", err)
		}

		if logEntry.Alert.Severity > 0 {
			timestamp, err := time.Parse(time.RFC3339, logEntry.Timestamp)
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
				log.Println(err)
			}
		}
	}

	return nil
}
