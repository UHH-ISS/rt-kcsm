package reader

import (
	"io"
	"rtkcsm/component/behaviour"
)

type AlertReader interface {
	ChannelAlerts(rtkcsm behaviour.RTKCSM, reader io.ReadCloser) error
}
