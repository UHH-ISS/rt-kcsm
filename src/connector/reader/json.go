package reader

import (
	"encoding/json"
	"io"
	"log"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
)

type JSONAlertReader struct{}

func (JAR *JSONAlertReader) ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error {
	defer reader.Close()

	decoder := json.NewDecoder(reader)
	// Expecting an array of structure.Alert in the JSON file
	var alerts structure.Alerts
	if err := decoder.Decode(&alerts); err != nil {
		log.Printf("error decoding file: %s", err)
		return err
	}

	// Send each alert to the channel
	for _, alert := range alerts {
		rtkcsm.Add(alert)
	}
	return nil
}
