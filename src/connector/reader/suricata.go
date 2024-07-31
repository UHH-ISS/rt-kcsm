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

type SuricataAlertReader struct{}

type suricataLogEntry struct {
	Timestamp    string        `json:"timestamp"`
	Source       string        `json:"src_ip"`
	Destination  string        `json:"dest_ip"`
	Alert        suricataAlert `json:"alert"`
	TruePositive bool          `json:"true_positive"`
}

type suricataAlert struct {
	Severity int `json:"severity"`
}

const maxSeverityLevel = 4

func (SR *SuricataAlertReader) ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error {
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
