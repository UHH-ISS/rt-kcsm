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

type OCSFDetectionFinding struct {
	ConfidenceID      int                    `json:"confidence_id"`
	EventTime         int                    `json:"time"`
	EvidenceArtifacts []OCSFEvidenceArtifact `json:"evidences"`
}

type OCSFEvidenceArtifact struct {
	SourcEndpoint       OCSFNetworkEndpoint `json:"src_endpoint"`
	DestinationEndpoint OCSFNetworkEndpoint `json:"dst_endpoint"`
}

type OCSFNetworkEndpoint struct {
	IP string `json:"ip"`
}

type OCSFAlertReader struct{}

func (ocsfAlertReader *OCSFAlertReader) ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var detectionFinding OCSFDetectionFinding
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &detectionFinding); err != nil {
			log.Printf("error decoding: %s (%s)\n", err, line)
		}

		seconds := int64(detectionFinding.EventTime)
		nanoseconds := int64((float64(detectionFinding.EventTime) - float64(seconds)) * 1_000_000_000)

		timestamp := time.Unix(seconds, nanoseconds)

		var sourceIP *structure.IPAddress
		var destinationIP *structure.IPAddress

		for _, evidence := range detectionFinding.EvidenceArtifacts {
			if evidence.DestinationEndpoint.IP != "" {
				ip := structure.ParseIPAddress(evidence.DestinationEndpoint.IP)
				destinationIP = &ip
			} else if evidence.SourcEndpoint.IP != "" {
				ip := structure.ParseIPAddress(evidence.SourcEndpoint.IP)
				sourceIP = &ip
			}

			if sourceIP != nil && destinationIP != nil {
				severity := float32(0)
				if detectionFinding.ConfidenceID == 1 { // Low
					severity = 0.33
				} else if detectionFinding.ConfidenceID == 2 { // Medium
					severity = 0.66
				} else if detectionFinding.ConfidenceID == 3 { // High
					severity = 1
				}

				alert := structure.Alert{
					Timestamp:     timestamp,
					SourceIP:      *sourceIP,
					DestinationIP: *destinationIP,
					Severity:      severity,
					TruePositive:  false,
				}

				rtkcsm.Add(alert)
			}
		}
	}

	return nil
}
