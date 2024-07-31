package structure

import (
	"encoding/json"
	"fmt"
	"time"
)

type DirectedRelationID string

type DirectedRelation struct {
	ID           DirectedRelationID
	DstNode      IPAddress
	SrcNode      IPAddress
	Timestamp    time.Time
	Stage        SimplifiedUKCStage
	Severity     float32
	TruePositive bool
}

func (r *DirectedRelation) Relevance() float32 {
	victimLabel := r.Stage.GetVictim()
	var victim IPAddress
	if victimLabel == VictimSource {
		victim = r.SrcNode
	} else {
		victim = r.DstNode
	}

	return r.Severity * float32(HostManager.GetHostRiskLevel(victim))
}

func (r *DirectedRelation) toDot(simplify bool) string {
	colorSrc := "blue"
	colorDst := "blue"
	sizeSrc := HostManager.GetHostRiskLevel(r.SrcNode)
	sizeDst := HostManager.GetHostRiskLevel(r.DstNode)

	if !r.SrcNode.IsPrivate() {
		colorSrc = "red"
	}

	if !r.DstNode.IsPrivate() {
		colorDst = "red"
	}

	src := r.SrcNode.String()
	dst := r.DstNode.String()

	if simplify {
		if !r.SrcNode.IsPrivate() {
			src = fmt.Sprintf("Internet-%s", dst)
		}

		if !r.DstNode.IsPrivate() {
			dst = fmt.Sprintf("Internet-%s", src)
		}
	}

	ukcStages := r.Stage.ToUKCStages()

	return fmt.Sprintf("\"%s\" [color=\"%s\" risk=%f]\n\"%s\" [color=\"%s\" risk=%f]\n\"%s\" -> \"%s\" [label=\"%v\" date=%d weight=%f];", src, colorSrc, sizeSrc, dst, colorDst, sizeDst, src, dst, ukcStages, r.Timestamp.Unix(), r.Severity)
}

type DirectedRelationJson[T Stage] struct {
	ID           DirectedRelationID `json:"id"`
	From         string             `json:"from"`
	To           string             `json:"to"`
	Stage        T                  `json:"stage"`
	Timestamp    string             `json:"timestamp"`
	Severity     float32            `json:"relevance"`
	TruePositive bool               `json:"true_positive"`
}

func (r *DirectedRelation) MarshalJSON() ([]byte, error) {
	return json.Marshal(DirectedRelationJson[SimplifiedUKCStage]{
		ID:           r.ID,
		From:         r.SrcNode.String(),
		To:           r.DstNode.String(),
		Stage:        r.Stage,
		Timestamp:    r.Timestamp.String(),
		Severity:     r.Severity,
		TruePositive: r.TruePositive,
	})
}

func (r *DirectedRelation) UnmarshalJSON(data []byte) error {
	jsonObject := DirectedRelationJson[SimplifiedUKCStage]{}
	err := json.Unmarshal(data, &jsonObject)
	if err != nil {
		return err
	}

	r.ID = jsonObject.ID
	r.SrcNode = ParseIPAddress(jsonObject.From)
	r.DstNode = ParseIPAddress(jsonObject.To)
	r.Stage = jsonObject.Stage
	r.TruePositive = jsonObject.TruePositive

	timestamp, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", jsonObject.Timestamp)
	if err != nil {
		return err
	}

	r.Timestamp = timestamp
	r.Severity = jsonObject.Severity

	return nil
}
