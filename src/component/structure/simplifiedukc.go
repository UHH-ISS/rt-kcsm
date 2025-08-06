package structure

import "errors"

type SimplifiedUKCStage int

const (
	Incoming SimplifiedUKCStage = iota
	SameZone
	DifferentZone
	Outgoing
	Host
	None
)

func NewSimplifiedUKCStageFromString(stage string) SimplifiedUKCStage {
	stageStringMap := map[string]SimplifiedUKCStage{
		"incoming":       Incoming,
		"same-zone":      SameZone,
		"different-zone": DifferentZone,
		"outgoing":       Outgoing,
		"host":           Host,
	}

	stageNumber, ok := stageStringMap[stage]
	if !ok {
		stageNumber = None
	}

	return stageNumber
}

var SimplifiedUkcStageWeights = map[SimplifiedUKCStage]float32{
	Incoming:      0.10,
	SameZone:      0.25,
	DifferentZone: 0.30,
	Outgoing:      0.35,
}

func (stage SimplifiedUKCStage) GetVictim() Direction {
	if stage == Outgoing {
		return Source
	} else {
		return Destination
	}
}

var ukcStageMapping = map[SimplifiedUKCStage][]UKCStage{
	Incoming:      {R, D1},
	SameZone:      {L, S, O},
	DifferentZone: {P, S, O},
	Outgoing:      {E, C2, D2},
}

func (stage SimplifiedUKCStage) Serialize() byte {
	return byte(stage)
}

func (stage SimplifiedUKCStage) GetWeight() float32 {
	return SimplifiedUkcStageWeights[stage]
}

func (stage SimplifiedUKCStage) ToUKCStages() []UKCStage {
	return ukcStageMapping[stage]
}

func IncomingStage(source IPAddress, destination IPAddress) bool {
	return !source.IsInternal() && destination.IsInternal()
}

func HostStage(source IPAddress, destination IPAddress) bool {
	return source.Equal(destination)
}

func InternalDifferentSubnetStage(source IPAddress, destination IPAddress) bool {
	return source.IsInternal() && destination.IsInternal() && !source.IsSameSubnet(destination)
}

func InternalSameSubnetStage(source IPAddress, destination IPAddress) bool {
	return source.IsInternal() && destination.IsInternal()
}

func OutgoingStage(source IPAddress, destination IPAddress) bool {
	return source.IsInternal() && !destination.IsInternal()
}

type SimplifiedUKCStageMapper struct{}

func NewSimplifiedUKCStageMapper() StageMapper[SimplifiedUKCStage] {
	return &SimplifiedUKCStageMapper{}
}

var errSimplifiedUkcStageMapperNotFound = errors.New("not found")

var simplifiedUkcPrecedingStages = map[SimplifiedUKCStage][]SimplifiedUKCStage{
	Incoming:      {},
	Host:          {Incoming, SameZone, Host, DifferentZone},
	SameZone:      {Incoming, SameZone, Host, DifferentZone, Outgoing},
	DifferentZone: {Incoming, SameZone, Host, DifferentZone, Outgoing},
	Outgoing:      {Incoming, SameZone, Host, DifferentZone, Outgoing},
}

func (SimplifiedUKCStageMapper) DetermineStage(alert Alert) (SimplifiedUKCStage, error) {
	source := alert.SourceIP
	destination := alert.DestinationIP

	if IncomingStage(source, destination) {
		return Incoming, nil
	} else if HostStage(source, destination) {
		return Host, nil
	} else if InternalDifferentSubnetStage(source, destination) {
		return DifferentZone, nil
	} else if InternalSameSubnetStage(source, destination) {
		return SameZone, nil
	} else if OutgoingStage(source, destination) {
		return Outgoing, nil
	} else {
		return None, errSimplifiedUkcStageMapperNotFound
	}
}

func (SimplifiedUKCStageMapper) GetPrecedingStages(stage SimplifiedUKCStage) []SimplifiedUKCStage {
	return simplifiedUkcPrecedingStages[stage]
}
