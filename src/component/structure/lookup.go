package structure

import (
	"time"
)

type LookupEntry [18]byte
type Direction byte

const SOURCE_ADDRESS Direction = 0b0000_0000
const DESTINATION_ADDRESS Direction = 0b0000_0001

func NewLookupEntry(ipAddress IPAddress, stage UKCStage) LookupEntry {
	ipAddressDirection := LookupEntry{}
	copy(ipAddressDirection[0:17], ipAddress[0:17])

	ipAddressDirection[17] = byte(stage)

	return ipAddressDirection
}

type TimeDirectedRelation struct {
	Earliest time.Time
}

type LookupTable struct {
	relations map[LookupEntry]map[GraphID]*TimeDirectedRelation
	graphs    map[GraphID]Set[LookupEntry]
}

func NewLookupTable() LookupTable {
	return LookupTable{
		relations: map[LookupEntry]map[GraphID]*TimeDirectedRelation{},
		graphs:    map[GraphID]Set[LookupEntry]{},
	}
}

func (l *LookupTable) GetGraphIds(relation *DirectedRelation) Set[GraphID] {
	graphIds := NewSet[GraphID]()

	if relation.SrcNode.IsPrivate() {
		for stage := range GetPreConditions(relation.Stage) {
			for graphID, r := range l.relations[NewLookupEntry(relation.SrcNode, stage)] {
				if !r.Earliest.After(relation.Timestamp) {
					graphIds.Append(graphID)
				}
			}
		}
	}

	return graphIds
}

func (l *LookupTable) MergeGraphs(oldGraphIds Set[GraphID], newGraphId GraphID) {
	for oldGraphId := range oldGraphIds {
		entries := l.graphs[oldGraphId]

		if l.graphs[newGraphId] == nil {
			l.graphs[newGraphId] = entries
		} else {
			l.graphs[newGraphId].Union(entries)
		}

		for entry := range entries {
			existingRelation := l.relations[entry][newGraphId]
			relation := l.relations[entry][oldGraphId]

			if existingRelation == nil {
				l.relations[entry][newGraphId] = relation
			} else if existingRelation.Earliest.After(relation.Earliest) {
				l.relations[entry][newGraphId].Earliest = relation.Earliest
			}

			if oldGraphId != newGraphId {
				delete(l.relations[entry], oldGraphId)
			}
		}

		if oldGraphId != newGraphId {
			delete(l.graphs, oldGraphId)
		}
	}
}

func (l *LookupTable) AddRelation(relation *DirectedRelation, graphId GraphID) {
	addresses := []IPAddress{}
	stageSets := [][]UKCStage{}
	stages := relation.Stage.ToUKCStages()

	if relation.SrcNode.IsPrivate() {
		addresses = append(addresses, relation.SrcNode)
		stageSets = append(stageSets, stages)
	}

	if relation.DstNode.IsPrivate() {
		addresses = append(addresses, relation.DstNode)
		stageSets = append(stageSets, stages)
	}

	for i := range len(addresses) {
		address := addresses[i]
		stages := stageSets[i]

		for _, stage := range stages {
			entry := NewLookupEntry(address, stage)

			if l.relations[entry] == nil {
				l.relations[entry] = map[GraphID]*TimeDirectedRelation{}
			}

			existingRelation := l.relations[entry][graphId]

			if existingRelation == nil {
				l.relations[entry][graphId] = &TimeDirectedRelation{
					Earliest: relation.Timestamp,
				}
			} else if existingRelation.Earliest.After(relation.Timestamp) {
				l.relations[entry][graphId].Earliest = relation.Timestamp
			}

			if l.graphs[graphId] == nil {
				l.graphs[graphId] = NewSet(entry)
			} else {
				l.graphs[graphId].Append(entry)
			}
		}
	}
}
