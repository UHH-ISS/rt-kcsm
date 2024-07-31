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

type SuricataTenzirAlertReader struct{}

func (SR *SuricataTenzirAlertReader) ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var logEntry suricataLogEntry
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &logEntry); err != nil {
			log.Printf("error decoding file: %s\n", err)
		}

		if logEntry.Alert.Severity > 0 {
			timestamp, err := time.Parse("2006-01-02T15:04:05.000000", logEntry.Timestamp)
			if err != nil {
				log.Printf("error parsing time: %s", err)
			}
			sourceIP := structure.ParseIPAddress(logEntry.Source)
			destinationIP := structure.ParseIPAddress(logEntry.Destination)
			severity := float32((maxSeverityLevel)-logEntry.Alert.Severity) / (maxSeverityLevel - 1)

			alert := structure.Alert{
				Timestamp:     timestamp,
				SourceIP:      sourceIP,
				DestinationIP: destinationIP,
				Severity:      severity,
				TruePositive:  logEntry.TruePositive,
			}

			rtkcsm.Add(alert)
		}
	}

	return nil
}
