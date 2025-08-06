package structure

import "cmp"

type Direction bool

const Source Direction = false
const Destination Direction = true

type Stage interface {
	cmp.Ordered
	GetVictim() Direction
	Serialize() byte
	GetWeight() float32
	ToUKCStages() []UKCStage
}
