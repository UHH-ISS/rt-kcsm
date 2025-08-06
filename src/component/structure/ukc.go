package structure

import "fmt"

type UKCStage int

const (
	R UKCStage = iota
	D1
	D2
	C2
	L
	S
	P
	E
	O
)

func (stage UKCStage) String() string {
	humanReadableNames := map[UKCStage]string{
		R:  "Reconnaissance",
		D1: "Delivery Phase 1",
		D2: "Delivery Phase 2",
		C2: "Command&Control",
		L:  "Lateral Movement",
		S:  "Discovery",
		P:  "Pivot",
		E:  "Exfiltration",
		O:  "Objectives",
	}

	return humanReadableNames[stage]
}

func (stage UKCStage) ToUKCStages() []UKCStage {
	panic("not implemented")
}

func (stage UKCStage) GetVictim() Direction {
	panic("not implemented")
}

func (stage UKCStage) GetWeight() float32 {
	panic("not implemented")
}

func (stage UKCStage) Serialize() byte {
	return byte(stage)
}

func (stage UKCStage) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", stage.String())), nil
}
