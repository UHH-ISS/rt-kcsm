package reader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"time"
)

type ZeekAlertReader[T structure.Stage, K structure.Stage] struct{}

type zeekAlert struct {
	UID         string  `json:"uid"`
	Timestamp   float64 `json:"ts"`
	Source      string  `json:"src"`
	Destination string  `json:"dst"`
	Label       string  `json:"label"`
	Message     string  `json:"msg"`
	Signature   string  `json:"note"`
}

func (zeekAlertReader *ZeekAlertReader[T, K]) ChannelAlerts(rtkcsm behaviour.RTKCSM[T, K], reader io.ReadCloser) error {
	defer reader.Close()

	zeekSignatureIdMap := map[string]uint32{}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var zeekAlert zeekAlert
		line := scanner.Bytes()

		if err := json.Unmarshal(line, &zeekAlert); err != nil {
			log.Printf("error decoding: %s (%s)\n", err, line)
		}

		seconds := int64(zeekAlert.Timestamp)
		nanoseconds := int64((zeekAlert.Timestamp - float64(seconds)) * 1_000_000_000)

		timestamp := time.Unix(seconds, nanoseconds)

		sourceIP := structure.ParseIPAddress(zeekAlert.Source)
		destinationIP := structure.ParseIPAddress(zeekAlert.Destination)

		signatureId, ok := zeekSignatureIdMap[zeekAlert.Signature]

		if !ok {
			length := len(zeekSignatureIdMap)
			if length <= math.MaxUint32 {
				signatureId = uint32(length)
				zeekSignatureIdMap[zeekAlert.Signature] = signatureId
			} else {
				return fmt.Errorf("exceeding signature id limit")
			}
		}

		alert := structure.Alert{
			Timestamp:     timestamp,
			SourceIP:      sourceIP,
			DestinationIP: destinationIP,
			Severity:      1,
			Confidence:    1,
			Cause:         zeekAlert.Message,
			SignatureId:   signatureId,
			Label:         zeekAlert.Label,
		}

		err := rtkcsm.AddAlert(alert)
		if err != nil {
			//log.Println(err)
		}
	}

	return nil
}
