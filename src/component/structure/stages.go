package structure

import "cmp"

type Victim bool

const VictimSource Victim = false
const VictimDestination Victim = true

type Stage interface {
	cmp.Ordered
	ToUKCStages() []UKCStage
	GetVictim() Victim
}
