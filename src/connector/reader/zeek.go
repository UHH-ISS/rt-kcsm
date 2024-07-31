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

type ZeekAlertReader struct{}

type zeekAlert struct {
	UID          string  `json:"uid"`
	Timestamp    float64 `json:"ts"`
	Source       string  `json:"src"`
	Destination  string  `json:"dst"`
	TruePositive bool    `json:"true_positive"`
}

func (zeekAlertReader *ZeekAlertReader) ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error {
	defer reader.Close()

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

		alert := structure.Alert{
			Timestamp:     timestamp,
			SourceIP:      sourceIP,
			DestinationIP: destinationIP,
			Severity:      1,
			TruePositive:  zeekAlert.TruePositive,
		}

		rtkcsm.Add(alert)
	}

	return nil
}
