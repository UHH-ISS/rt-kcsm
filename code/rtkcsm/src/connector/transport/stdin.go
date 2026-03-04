package transport

import (
	"os"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
)

type StdinTransport[T structure.Stage, K structure.Stage] struct{}

func (transport *StdinTransport[T, K]) Start(rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error {
	return reader.ChannelAlerts(rtkcsm, os.Stdin)
}
