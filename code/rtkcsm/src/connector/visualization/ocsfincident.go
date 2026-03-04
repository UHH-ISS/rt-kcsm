package visualization

import (
	"rtkcsm/component/structure"
	"time"
)

type OCSFSeverity int

const OCSFSeverityUnknown OCSFSeverity = 0

const OCSFSeverityInformational OCSFSeverity = 1
const OCSFSeverityLow OCSFSeverity = 2
const OCSFSeverityMedium OCSFSeverity = 3
const OCSFSeverityHigh OCSFSeverity = 4
const OCSFSeverityCritical OCSFSeverity = 5
const OCSFSeverityFatal OCSFSeverity = 6

type OCSFType int

const OCSFTypeIncidentFindingCreate OCSFType = 200501
const OCSFTypeDetectionFindingCreate OCSFType = 200401

type OCSFClass int

const OCSFClassIncidentFinding OCSFClass = 2005

type OCSFActivity int

const OCSFActivityCreate OCSFActivity = 1

type OCSFCategory int

const OCSFCategoryFindings OCSFCategory = 2

type OCSFStatus int

const OCSFStatusNew OCSFStatus = 1

type OCSFBase struct {
	Type      OCSFType     `json:"type_uid"`
	Class     OCSFClass    `json:"class_uid"`
	Activity  OCSFActivity `json:"activity_id"`
	Category  OCSFCategory `json:"category_uid"`
	Timestamp int          `json:"time"`
	Metadata  OCSFMetadata `json:"metadata"`
	Severity  OCSFSeverity `json:"severity_id"`
	Status    OCSFStatus   `json:"status_id"`
}

type OCSFMetadata struct {
	Version string      `json:"version"`
	Product OCSFProduct `json:"product"`
}

type OCSFProduct struct {
	Name string `json:"name"`
}

type OCSFKillChainPhaseType int

const OCSFKillChainPhaseReconnaissance OCSFKillChainPhaseType = 1
const OCSFKillChainPhaseDelivery OCSFKillChainPhaseType = 3
const OCSFKillChainPhaseCommandAndControl OCSFKillChainPhaseType = 6
const OCSFKillChainPhaseActionsOnObjectives OCSFKillChainPhaseType = 7
const OCSFKillChainPhaseOther OCSFKillChainPhaseType = 99

type OCSFKillChainPhase struct {
	Type OCSFKillChainPhaseType `json:"phase_id"`
	Name string                 `json:"phase"`
}

type OCSFObservableType int

const OCSFObservableTypeIPAddress OCSFObservableType = 2

type OCSFObservable struct {
	Name  string             `json:"name"`
	Type  OCSFObservableType `json:"type_id"`
	Value string             `json:"value"`
}

type OCSFRelatedEvent struct {
	ID                 string               `json:"uid"`
	Type               OCSFType             `json:"type_uid"`
	KillChainPhases    []OCSFKillChainPhase `json:"kill_chain"`
	FirstSeenTimestamp int                  `json:"fist_seen_time"`
	Count              int                  `json:"count"`
	Description        string               `json:"desc"`
	Observables        []OCSFObservable     `json:"observables"`
}

type OCSFFindingInformation struct {
	ID            string             `json:"uid"`
	RelatedEvents []OCSFRelatedEvent `json:"related_events"`
}

type OCSFIncidentFinding struct {
	OCSFBase
	FindingInformations []OCSFFindingInformation `json:"finding_info_list"`
}

var ukcStagesToOCSFKillChainPhases = map[structure.UKCStage]OCSFKillChainPhaseType{
	structure.C2: OCSFKillChainPhaseCommandAndControl,
	structure.R:  OCSFKillChainPhaseReconnaissance,
	structure.O:  OCSFKillChainPhaseActionsOnObjectives,
}

func FromGraphToOCSFIncidentFinding[T structure.Stage](graph *structure.PreComputedGraph[T]) OCSFIncidentFinding {
	var severity OCSFSeverity
	severityLevel := int(graph.ComputedRelevance * 6)
	switch severityLevel {
	case 1:
		severity = OCSFSeverityInformational
	case 2:
		severity = OCSFSeverityLow
	case 3:
		severity = OCSFSeverityMedium
	case 4:
		severity = OCSFSeverityHigh
	case 5:
		severity = OCSFSeverityCritical
	default:
		if severity <= 0 {
			severity = OCSFSeverityUnknown
		} else {
			severity = OCSFSeverityFatal
		}
	}

	findings := []OCSFFindingInformation{}

	for _, relation := range graph.PreComputedDirectedRelations {
		killChainPhases := []OCSFKillChainPhase{}

		for _, stage := range relation.ConfirmedStages {
			killChainPhaseType, ok := ukcStagesToOCSFKillChainPhases[stage]
			if !ok {
				killChainPhaseType = OCSFKillChainPhaseOther
			}
			killChainPhases = append(killChainPhases, OCSFKillChainPhase{
				Type: killChainPhaseType,
				Name: stage.String(),
			})
		}

		findings = append(findings, OCSFFindingInformation{
			ID: relation.ID,
			RelatedEvents: []OCSFRelatedEvent{
				{
					ID:                 relation.ID,
					Type:               OCSFTypeDetectionFindingCreate,
					KillChainPhases:    killChainPhases,
					FirstSeenTimestamp: int(relation.Timestamp),
					Count:              relation.Count,
					Description:        relation.Cause,
					Observables: []OCSFObservable{
						{
							Name:  "source",
							Type:  OCSFObservableTypeIPAddress,
							Value: relation.From,
						},
						{
							Name:  "destination",
							Type:  OCSFObservableTypeIPAddress,
							Value: relation.To,
						},
					},
				},
			},
		})
	}

	return OCSFIncidentFinding{
		OCSFBase: OCSFBase{
			Type:      OCSFTypeIncidentFindingCreate,
			Class:     OCSFClassIncidentFinding,
			Activity:  OCSFActivityCreate,
			Category:  OCSFCategoryFindings,
			Timestamp: int(time.Now().UnixMilli()),
			Metadata: OCSFMetadata{
				Version: "1.0",
				Product: OCSFProduct{
					Name: "RT-KCSM",
				},
			},
			Severity: severity,
			Status:   OCSFStatusNew,
		},
		FindingInformations: findings,
	}
}
