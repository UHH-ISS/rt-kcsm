package structure

type UKCStage int

const (
	R UKCStage = iota
	D1
	D2
	C
	H
	L
	S
	P
	E
	O
)

func (stage UKCStage) String() string {
	humanReadableNames := map[UKCStage]string{
		R:  "R",
		D1: "D1",
		D2: "D2",
		C:  "C2",
		H:  "H",
		L:  "L",
		S:  "S",
		P:  "P",
		E:  "E",
		O:  "O",
	}

	return humanReadableNames[stage]
}

func (stage UKCStage) ToUKCStages() []UKCStage {
	return []UKCStage{stage}
}

func (stage UKCStage) GetVictim() Victim {
	if stage == D1 || stage == D2 || stage == C || stage == S || stage == E || stage == O {
		return VictimSource
	} else {
		return VictimDestination
	}
}

var PreConditions = map[UKCStage][]UKCStage{
	R:  {},
	D1: {R},
	D2: {D1, H},
	C:  {D1, D2, C, L, H},
	L:  {D1, D2, C, L, E, O, S, H},
	P:  {D1, D2, C, L, H},
	S:  {D1, D2, C, L, P, H},
	E:  {D1, D2, C, L, E, O, H},
	O:  {D1, D2, C, L, E, O, H},
	H:  {L, C, D1, D2, H},
}

func GetPreConditions[T Stage](stages ...T) Set[UKCStage] {
	results := NewSet[UKCStage]()
	for _, stage := range stages {
		for _, ukcstage := range stage.ToUKCStages() {
			results.Append(PreConditions[ukcstage]...)
		}
	}
	return results
}
