package structure

import (
	"time"
)

type DirectedRelationID string

type DirectedRelation[T Stage] struct {
	DstNode   IPAddress `json:"from"`
	SrcNode   IPAddress `json:"to"`
	Timestamp time.Time
	MetaStage T

	Severity    float32
	Confidence  float32
	Cause       string
	SignatureId uint32
	Labels      []string
}

type DirectedRelationGroup[T Stage] struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	MetaStage   T        `json:"stage"`
	Timestamp   int64    `json:"timestamp"`
	Severity    float32  `json:"severity"`
	Confidence  float32  `json:"confidence"`
	SignatureId uint32   `json:"signature_id"`
	Cause       string   `json:"cause"`
	Labels      []string `json:"labels"`
	Count       int      `json:"count"`
}
