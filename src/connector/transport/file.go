package transport

import (
	"os"
	"rtkcsm/component/behaviour"
	"rtkcsm/connector/reader"
)

type FileTransport struct {
	FilePath string
}

func (transport *FileTransport) Start(rtkcsm behaviour.RTKCSM, reader reader.AlertReader) error {
	file, err := os.Open(transport.FilePath)
	if err != nil {
		return err
	}

	return reader.ChannelAlerts(rtkcsm, file)
}
