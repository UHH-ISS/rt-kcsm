package behaviour

import (
	"io"
	"rtkcsm/component/structure"
)

type RTKCSM interface {
	Add(alert structure.Alert)
	ImportSave(reader io.Reader) error
	GetGraphList(page int) structure.GraphInformationList
	GetGraph(id structure.GraphID) *structure.Graph
	GetHostRisks() []structure.HostRisk
	AddHostRisk(address structure.IPAddress, riskLevel structure.RiskLevel)
	DeleteHostRisk(address structure.IPAddress)
	Stop(exportFormats []structure.ExportFormat)
	Reset()
	GetEventManager() *structure.EventManager
}
