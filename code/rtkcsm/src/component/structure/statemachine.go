package structure

import "rtkcsm/component/structure/set"

type StateMachine[T Stage, K Stage] interface {
	GetPrecedingStages(stage T) []K
	GetCurrentStateStages(stage T) []K
}

var ukcPrecedingStages = map[UKCStage][]UKCStage{
	R:  {},
	D1: {R},
	D2: {R, D1},
	C2: {L, P, C2, D1, D2},
	L:  {L, P, S, C2, D1, D2},
	P:  {L, P, C2, D1, D2},
	S:  {L, S, C2, D1, D2},
	E:  {L, P, D1, D2, O, E},
	O:  {L, P, D1, D2, O, E},
}

type UKCStateMachine[T Stage] struct{}

func NewUKCStateMachine[T Stage]() StateMachine[T, UKCStage] {
	return &UKCStateMachine[T]{}
}

func (s UKCStateMachine[T]) GetPrecedingStages(stage T) []UKCStage {
	set := set.NewSet[UKCStage]()
	for _, stage := range stage.ToUKCStages() {
		set.Append(ukcPrecedingStages[stage]...)
	}

	return set.ToSlice()
}

func (s UKCStateMachine[T]) GetCurrentStateStages(stage T) []UKCStage {
	return stage.ToUKCStages()
}
