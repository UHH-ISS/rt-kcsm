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
	TypeUID            int                    `json:"type_uid"`
	ConfidenceID       int                    `json:"confidence_id"`
	SeverityId         int                    `json:"severity_id"`
	EventTime          int                    `json:"time"`
	EvidenceArtifacts  []OCSFEvidenceArtifact `json:"evidences"`
	Unmapped           suricataLogEntry       `json:"unmapped"`
	FindingInformation OCSFFindingInformation `json:"finding_info"`
}

type OCSFFindingInformation struct {
	Description string `json:"desc"`
}

type OCSFEvidenceArtifact struct {
	SourceEndpoint      OCSFNetworkEndpoint `json:"src_endpoint"`
	DestinationEndpoint OCSFNetworkEndpoint `json:"dst_endpoint"`
}

type OCSFNetworkEndpoint struct {
	IP string `json:"ip"`
}

var DetectionFindingCreateTypeUID = 200401

type OCSFAlertReader[T structure.Stage, K structure.Stage] struct{}

func (ocsfAlertReader *OCSFAlertReader[T, K]) ChannelAlerts(rtkcsm behaviour.RTKCSM[T, K], reader io.ReadCloser) error {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var detectionFinding OCSFDetectionFinding
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &detectionFinding); err != nil {
			log.Printf("error decoding: %s (%s)\n", err, line)
		}

		// Check if detection finding is created
		if detectionFinding.TypeUID == DetectionFindingCreateTypeUID {
			timestamp := time.UnixMilli(int64(detectionFinding.EventTime))

			var sourceIP *structure.IPAddress
			var destinationIP *structure.IPAddress

			for _, evidence := range detectionFinding.EvidenceArtifacts {
				if evidence.DestinationEndpoint.IP != "" {
					ip := structure.ParseIPAddress(evidence.DestinationEndpoint.IP)
					destinationIP = &ip
				} else if evidence.SourceEndpoint.IP != "" {
					ip := structure.ParseIPAddress(evidence.SourceEndpoint.IP)
					sourceIP = &ip
				}

				if sourceIP == nil {
					sourceIP = destinationIP
				} else if destinationIP == nil {
					destinationIP = sourceIP
				}

				if sourceIP != nil && destinationIP != nil {
					confidence := float32(0)
					switch detectionFinding.ConfidenceID {
					case 1: // Low
						confidence = 0.33
					case 2: // Medium
						confidence = 0.66
					default: // High
						confidence = 1
					}

					severity := float32(detectionFinding.SeverityId)
					if severity >= 1 && severity <= 6 { // Severity ids are from 1 to 6
						severity = severity / 6
					} else { // If unknown or other we map it to 1
						severity = 1
					}

					alert := structure.Alert{
						Timestamp:     timestamp,
						SourceIP:      *sourceIP,
						DestinationIP: *destinationIP,
						Severity:      severity,
						Confidence:    confidence,
						SignatureId:   detectionFinding.Unmapped.Alert.SignatureId,
						Cause:         detectionFinding.FindingInformation.Description,
						Label:         "", // Not applicable for production as it is used for testing
					}

					err := rtkcsm.AddAlert(alert)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	return nil
}
