package reader

import (
	"io"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
)

type AlertReader[T structure.Stage, K structure.Stage] interface {
	ChannelAlerts(rtkcsm behaviour.RTKCSM[T, K], reader io.ReadCloser) error
}
