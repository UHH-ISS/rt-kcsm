package structure

import (
	"errors"
	"sort"
)

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
	Host:          1.0,
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
	Host:          {X},
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

func IncomingStage(source *IPAddress, destination *IPAddress) bool {
	return !source.IsInternal() && destination.IsInternal()
}

func HostStage(source *IPAddress, destination *IPAddress) bool {
	return source.Equal(destination)
}

func InternalStage(source *IPAddress, destination *IPAddress) bool {
	return source.IsInternal() && destination.IsInternal()
}

func SameNetwork(sourceNetwork *IPNetwork, destinationNetwork *IPNetwork) bool {
	return sourceNetwork != nil && destinationNetwork != nil && sourceNetwork.Equal(destinationNetwork)
}

func OutgoingStage(source *IPAddress, destination *IPAddress) bool {
	return source.IsInternal() && !destination.IsInternal()
}

type SimplifiedUKCStageMapper struct {
	internalNetworks IPNetworks
}

func NewSimplifiedUKCStageMapper(internalNetworkStrings ...string) StageMapper[SimplifiedUKCStage] {
	internalNetworks := IPNetworks{}

	if len(internalNetworkStrings) == 0 {
		internalNetworks = append(internalNetworks, ParseIPNetwork("192.168.0.0/16"))
		internalNetworks = append(internalNetworks, ParseIPNetwork("172.16.0.0/12"))
		internalNetworks = append(internalNetworks, ParseIPNetwork("10.0.0.0/8"))
		internalNetworks = append(internalNetworks, ParseIPNetwork("fc00::/7"))
	} else {
		for _, internalNetwork := range internalNetworkStrings {
			internalNetworks = append(internalNetworks, ParseIPNetwork(internalNetwork))
		}
	}

	sort.Sort(sort.Reverse(internalNetworks))

	return &SimplifiedUKCStageMapper{
		internalNetworks: internalNetworks,
	}
}

var errSimplifiedUkcStageMapperNotFound = errors.New("not found")

var simplifiedUkcPrecedingStages = map[SimplifiedUKCStage][]SimplifiedUKCStage{
	Incoming:      {},
	Host:          {Incoming, SameZone, Host, DifferentZone},
	SameZone:      {Incoming, SameZone, Host, DifferentZone, Outgoing, Host},
	DifferentZone: {Incoming, SameZone, Host, DifferentZone, Outgoing, Host},
	Outgoing:      {Incoming, SameZone, Host, DifferentZone, Outgoing, Host},
}

func (mapper *SimplifiedUKCStageMapper) DetermineStage(alert *Alert) (SimplifiedUKCStage, error) {
	source := &alert.SourceIP
	var sourceNetwork *IPNetwork = nil
	destination := &alert.DestinationIP
	var destinationNetwork *IPNetwork = nil

	for _, internalNetwork := range mapper.internalNetworks {
		if internalNetwork.HasIpAddress(source) {
			source.SetInternal(true)
			sourceNetwork = &internalNetwork
			break
		}
	}

	for _, internalNetwork := range mapper.internalNetworks {
		if internalNetwork.HasIpAddress(destination) {
			destination.SetInternal(true)
			destinationNetwork = &internalNetwork
			break
		}
	}

	if IncomingStage(source, destination) {
		return Incoming, nil
	} else if HostStage(source, destination) {
		return Host, nil
	} else if InternalStage(source, destination) {
		if SameNetwork(sourceNetwork, destinationNetwork) {
			return SameZone, nil
		} else {
			return DifferentZone, nil
		}
	} else if OutgoingStage(source, destination) {
		return Outgoing, nil
	} else {
		return None, errSimplifiedUkcStageMapperNotFound
	}
}

func (mapper SimplifiedUKCStageMapper) GetPrecedingStages(stage SimplifiedUKCStage) []SimplifiedUKCStage {
	return simplifiedUkcPrecedingStages[stage]
}
