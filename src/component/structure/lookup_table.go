package structure

import (
	"rtkcsm/component/structure/set"
	"rtkcsm/component/structure/timebucket"
)

type LookupEntry [IP_ADDRESS_LENGTH + 1]byte

func NewLookupEntry[T Stage](ipAddress IPAddress, stage T) LookupEntry {
	ipAddressDirection := LookupEntry{}
	copy(ipAddressDirection[0:IP_ADDRESS_LENGTH], ipAddress[0:IP_ADDRESS_LENGTH]) // 17 bytes
	ipAddressDirection[IP_ADDRESS_LENGTH] = stage.Serialize()                     // 1 byte

	return ipAddressDirection
}

type LookupTable[T Stage, K Stage] struct {
	relations    map[LookupEntry]*timebucket.TimeBucketIndex[GraphID]
	stateMachine StateMachine[T, K]
}

func NewLookupTable[T Stage, K Stage](stateMachine StateMachine[T, K]) LookupTable[T, K] {
	return LookupTable[T, K]{
		relations:    map[LookupEntry]*timebucket.TimeBucketIndex[GraphID]{},
		stateMachine: stateMachine,
	}
}

func (l *LookupTable[T, K]) SearchRelations(relation *EnrichedAlert[T]) set.Set[GraphID] {
	graphIds := set.NewSet[GraphID]()
	if relation.Alert.SourceIP.IsInternal() {
		for _, precedingStage := range l.stateMachine.GetPrecedingStages(relation.MetaStage) {
			entry := NewLookupEntry(relation.Alert.SourceIP, precedingStage)
			bucketIndex := l.relations[entry]
			if bucketIndex != nil {
				values := bucketIndex.GetAllBefore(relation.Timestamp)
				graphIds.Append(values...)
			}
		}
	}

	return graphIds
}

func (l *LookupTable[T, K]) AddRelation(relation *DirectedRelation[T], graphID GraphID, graph *Graph[T, K]) {
	addresses := []IPAddress{}

	if relation.SrcNode.IsInternal() {
		addresses = append(addresses, relation.SrcNode)
	}

	if relation.DstNode.IsInternal() {
		addresses = append(addresses, relation.DstNode)
	}

	stages := l.stateMachine.GetCurrentStateStages(relation.MetaStage)

	for i := range len(addresses) {
		address := addresses[i]

		for _, stage := range stages {
			entry := NewLookupEntry(address, stage)

			bucketIndex := l.relations[entry]

			if bucketIndex == nil {
				bucketIndex = timebucket.NewTimeBucketIndex[GraphID]()
				l.relations[entry] = bucketIndex
			}

			bucket := graph.GetBucket(address, stage)

			if bucket != nil {
				if bucket.Get(graphID).After(relation.Timestamp) {
					bucket.Delete(graphID)
					bucket = bucketIndex.Insert(graphID, relation.Timestamp)
				}
			} else {
				bucket = bucketIndex.Insert(graphID, relation.Timestamp)
			}

			graph.StoreBucket(address, stage, bucket)
		}
	}
}
