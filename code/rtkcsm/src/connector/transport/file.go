package transport

import (
	"os"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
)

type FileTransport[T structure.Stage, K structure.Stage] struct {
	FilePath string
}

func (transport *FileTransport[T, K]) Start(rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
	file, err := os.Open(transport.FilePath)
	if err != nil {
		return err
	}

	return reader.ChannelAlerts(rtkcsm, file)
}
