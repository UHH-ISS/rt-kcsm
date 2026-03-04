package behaviour

import (
	"io"
	"rtkcsm/component/structure"
)

type RTKCSM[T structure.Stage, K structure.Stage] interface {
	AddAlert(alert structure.Alert) error
	GetGraphList(limit int) structure.GraphInformationList
	GetGraph(id structure.GraphID) *structure.Graph[T, K]
	GetHostRisks() []structure.HostRisk
	AddHostRisk(address structure.IPAddress, riskLevel structure.RiskLevel)
	DeleteHostRisk(address structure.IPAddress)
	ImportGraphs(reader io.Reader) error
	ExportGraphs(writer io.Writer) (int, error)
	Reset()
}
