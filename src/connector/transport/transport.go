package transport

import (
	"rtkcsm/component/behaviour"
	"rtkcsm/connector/reader"
)

type Transport interface {
	Start(rtkcsm behaviour.RTKCSM, reader reader.AlertReader) error
}
