package visualization

import "rtkcsm/component/structure"

type NodeRelationEventData struct {
	Address   string  `json:"address"`
	IsPrivate bool    `json:"is_private"`
	Risk      float32 `json:"risk"`
}

type RelationEventData struct {
	From      NodeRelationEventData `json:"from"`
	To        NodeRelationEventData `json:"to"`
	Stages    string                `json:"stages"`
	Timestamp string                `json:"timestamp"`
	Severity  float32               `json:"severity"`
}

type RelationEvent struct {
	Relation       RelationEventData `json:"relation"`
	GraphRelevance float32           `json:"relevance"`
	GraphID        structure.GraphID `json:"id"`
}
