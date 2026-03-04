package transport

import (
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
)

type Transport[T structure.Stage, K structure.Stage] interface {
	Start(rtkcsm behaviour.RTKCSM[T, K], reader reader.AlertReader[T, K]) error
}
