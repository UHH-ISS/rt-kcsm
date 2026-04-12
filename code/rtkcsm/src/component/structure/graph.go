package structure

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"rtkcsm/component/structure/timebucket"
	"sort"
	"sync"
	"time"
)

type GraphID int

var nextGraphID GraphID = 0

func NextGraphID() GraphID {
	nextGraphID += 1
	return nextGraphID
}

type GraphInformationList struct {
	Graphs []GraphInformation `json:"graphs"`
	Count  int                `json:"count"`
}

const OPTIMIZED_DIRECTED_RELATION_ID_LENGTH = 2*IP_ADDRESS_LENGTH + 4*3

type OptimizedDirectedRelationID [OPTIMIZED_DIRECTED_RELATION_ID_LENGTH]byte

func NewOptimizedDirectedRelationID(src IPAddress, dst IPAddress, severity float32, confidence float32, signatureId uint32) OptimizedDirectedRelationID {
	optimizedDirectedRelationID := [OPTIMIZED_DIRECTED_RELATION_ID_LENGTH]byte{}

	copy(optimizedDirectedRelationID[0:17], src[:])
	copy(optimizedDirectedRelationID[17:34], dst[:])

	binary.BigEndian.PutUint32(optimizedDirectedRelationID[34:38], math.Float32bits(severity))
	binary.BigEndian.PutUint32(optimizedDirectedRelationID[38:42], math.Float32bits(confidence))
	binary.BigEndian.PutUint32(optimizedDirectedRelationID[42:46], signatureId)

	return optimizedDirectedRelationID
}

func (o OptimizedDirectedRelationID) GetSrc() IPAddress {
	address := IPAddress{}
	copy(address[:], o[0:IP_ADDRESS_LENGTH])
	return address
}

func (o OptimizedDirectedRelationID) GetDst() IPAddress {
	address := IPAddress{}
	copy(address[:], o[IP_ADDRESS_LENGTH:IP_ADDRESS_LENGTH*2])
	return address
}

func (o OptimizedDirectedRelationID) Severity() float32 {
	offset := 34
	f := math.Float32frombits(uint32(o[offset])<<24 | uint32(o[offset+1])<<16 | uint32(o[offset+2])<<8 | uint32(o[offset+3]))
	return f
}

func (o OptimizedDirectedRelationID) Confidence() float32 {
	offset := 38
	f := math.Float32frombits(uint32(o[offset])<<24 | uint32(o[offset+1])<<16 | uint32(o[offset+2])<<8 | uint32(o[offset+3]))
	return f
}

func (o OptimizedDirectedRelationID) Relevance() float32 {
	return o.Confidence() * o.Severity()
}

func (o OptimizedDirectedRelationID) GetSignatureId() uint32 {
	return binary.BigEndian.Uint32(o[42:46])
}

func (o OptimizedDirectedRelationID) String() string {
	return string(bytes.Trim(o[:], "\x00"))
}

var allLabels = map[string]uint64{}

type OptimizedDirectedRelation[T Stage] struct {
	MetaStage T
	Timestamp time.Time
	Labels    uint64
	Count     int
	Cause     string
}

func NewOptimizedDirectedRelation[T Stage](relation DirectedRelation[T]) (OptimizedDirectedRelation[T], OptimizedDirectedRelationID) {

	return OptimizedDirectedRelation[T]{
		MetaStage: relation.MetaStage,
		Timestamp: relation.Timestamp,
		Cause:     relation.Cause,
		Count:     0,
	}, NewOptimizedDirectedRelationID(relation.SrcNode, relation.DstNode, relation.Severity, relation.Confidence, relation.SignatureId)
}

func (r *OptimizedDirectedRelation[T]) AddLabel(labels ...string) {
	for _, label := range labels {
		bit, ok := allLabels[label]
		if !ok {
			bit = uint64(1 << len(allLabels))
			allLabels[label] = bit
		}

		r.Labels |= bit
	}
}

func (r *OptimizedDirectedRelation[T]) GetLabels() []string {
	matchingLabels := []string{}

	for name, i := range allLabels {
		if r.Labels&i == i {
			matchingLabels = append(matchingLabels, name)
		}
	}

	return matchingLabels
}

func (r *OptimizedDirectedRelation[T]) Relevance(id OptimizedDirectedRelationID) float32 {
	victimLabel := r.MetaStage.GetVictim()
	var victim IPAddress
	if victimLabel == Source {
		victim = id.GetSrc()
	} else {
		victim = id.GetDst()
	}

	return id.Relevance() * float32(HostManager.GetHostRiskLevel(victim))
}

type ReverseLookupEntry[T Stage] struct {
	Buckets             map[T]*timebucket.Bucket[GraphID]
	HasLateralMovement  bool
	HasOutgoingActivity bool
}

type GraphInformation struct {
	ID        GraphID `json:"id"`
	Relevance float32 `json:"relevance"`
}

type ReverseLookupStoreRequest[T Stage] struct {
	SrcBucket  *timebucket.Bucket[GraphID]
	DstBucket  *timebucket.Bucket[GraphID]
	MetaStages []T
}

type ReverseLookupGetRequest[T Stage] struct {
	Address   IPAddress
	MetaStage T
}

type StoreDirectedRelationRequest[T Stage] struct {
	Relation                  DirectedRelation[T]
	ReverseLookupStoreRequest ReverseLookupStoreRequest[T]
}

type Graph[T Stage, K Stage] struct {
	Relations         map[OptimizedDirectedRelationID]OptimizedDirectedRelation[T] `json:"relations"`
	reverseLookup     map[IPAddress]ReverseLookupEntry[K]
	relevances        map[T]float32
	relationsMutex    *sync.RWMutex
	ComputedRelevance float32 `json:"computed_relevance"`
}

func NewGraph[T Stage, K Stage]() *Graph[T, K] {
	graph := Graph[T, K]{
		Relations:      map[OptimizedDirectedRelationID]OptimizedDirectedRelation[T]{},
		relevances:     map[T]float32{},
		relationsMutex: &sync.RWMutex{},
		reverseLookup:  map[IPAddress]ReverseLookupEntry[K]{},
	}

	return &graph
}

func (g *Graph[T, K]) Relevance() float32 {
	return g.ComputedRelevance
}

func (g *Graph[T, K]) GetRelations() []DirectedRelation[T] {
	relations := []DirectedRelation[T]{}
	for id, r := range g.Relations {
		relations = append(relations, DirectedRelation[T]{
			SrcNode:     id.GetSrc(),
			DstNode:     id.GetDst(),
			Severity:    id.Severity(),
			Confidence:  id.Confidence(),
			Timestamp:   r.Timestamp,
			MetaStage:   r.MetaStage,
			Cause:       r.Cause,
			SignatureId: id.GetSignatureId(),
		})
	}

	return relations
}

func (g *Graph[T, K]) Append(r EnrichedAlert[T]) DirectedRelation[T] {
	relation := DirectedRelation[T]{
		SrcNode:     r.SourceIP,
		DstNode:     r.DestinationIP,
		Severity:    r.Severity,
		Confidence:  r.Confidence,
		SignatureId: r.SignatureId,
		Timestamp:   r.Timestamp,
		MetaStage:   r.MetaStage,
		Cause:       r.Cause,
		Labels:      []string{r.Label},
	}
	g.append(relation)

	return relation
}

func (g *Graph[T, K]) append(r DirectedRelation[T]) {
	g.relationsMutex.Lock()
	defer g.relationsMutex.Unlock()

	relation, id := NewOptimizedDirectedRelation(r)
	if existingRelation, ok := g.Relations[id]; ok {
		relation = existingRelation
	}

	relation.Count += 1
	relation.AddLabel(r.Labels...)

	g.Relations[id] = relation
	relationRelevance := relation.Relevance(id)
	existingMaxStageRelevance := g.relevances[relation.MetaStage]

	if existingMaxStageRelevance < relationRelevance {
		g.relevances[relation.MetaStage] = relationRelevance
		g.ComputedRelevance += (relationRelevance - existingMaxStageRelevance) * relation.MetaStage.GetWeight()
	}

	g.updatePredecessor(r.SrcNode, r.SrcNode.IsInternal() && r.DstNode.IsInternal() && !r.DstNode.Equal(r.SrcNode), r.SrcNode.IsInternal())
	g.updatePredecessor(r.DstNode, false, false)
}

func (g *Graph[T, K]) updatePredecessor(address IPAddress, hasLateralMovement bool, hasOutgoingActivity bool) {
	if address.IsInternal() {
		entry, ok := g.reverseLookup[address]
		if !ok {
			entry = ReverseLookupEntry[K]{
				Buckets:             map[K]*timebucket.Bucket[GraphID]{},
				HasLateralMovement:  hasLateralMovement,
				HasOutgoingActivity: hasOutgoingActivity,
			}
		} else {
			if !entry.HasLateralMovement {
				entry.HasLateralMovement = hasLateralMovement
			}

			if !entry.HasOutgoingActivity {
				entry.HasOutgoingActivity = hasOutgoingActivity
			}
		}

		g.reverseLookup[address] = entry
	}
}

func (g *Graph[T, K]) StoreBucket(address IPAddress, stage K, bucket *timebucket.Bucket[GraphID]) {
	entry, ok := g.reverseLookup[address]
	if !ok {
		entry = ReverseLookupEntry[K]{
			Buckets: map[K]*timebucket.Bucket[GraphID]{},
		}
	}
	entry.Buckets[stage] = bucket
	g.reverseLookup[address] = entry
}

func (g *Graph[T, K]) GetBucket(address IPAddress, stage K) *timebucket.Bucket[GraphID] {
	if entry, ok := g.reverseLookup[address]; ok {
		return entry.Buckets[stage]
	}

	return nil
}

func (g *Graph[T, K]) GetBuckets(address IPAddress) map[K]*timebucket.Bucket[GraphID] {
	if entry, ok := g.reverseLookup[address]; ok {
		return entry.Buckets
	}

	return map[K]*timebucket.Bucket[GraphID]{}
}

func (g *Graph[T, K]) RecomputeRelevance() float32 {
	g.relevances = map[T]float32{}
	g.ComputedRelevance = 0

	for id, relation := range g.Relations {
		relationRelevance := relation.Relevance(id)
		existingMaxStageRelevance := g.relevances[relation.MetaStage]

		if existingMaxStageRelevance < relationRelevance {
			g.relevances[relation.MetaStage] = relationRelevance
			g.ComputedRelevance += (relationRelevance - existingMaxStageRelevance) * relation.MetaStage.GetWeight()
		}
	}

	return g.ComputedRelevance
}

func (g *Graph[T, K]) Merge(otherGraph *Graph[T, K], oldGraphID GraphID, newGraphId GraphID) {
	g.relationsMutex.Lock()
	defer g.relationsMutex.Unlock()
	otherGraph.relationsMutex.RLock()
	defer otherGraph.relationsMutex.RUnlock()
	for id, relation := range otherGraph.Relations {
		if existingRelation, ok := g.Relations[id]; ok {
			relation.Count += existingRelation.Count
			if existingRelation.Timestamp.Before(relation.Timestamp) {
				relation.Timestamp = existingRelation.Timestamp
			}
			relation.AddLabel(existingRelation.GetLabels()...)
		}

		g.Relations[id] = relation

		src := id.GetSrc()
		dst := id.GetDst()

		for stage, oldGraphBucket := range otherGraph.GetBuckets(src) {
			newGraphBucket := g.GetBucket(src, stage)
			oldTimestamp := oldGraphBucket.Get(newGraphId)

			if newGraphBucket == nil || newGraphBucket.Get(newGraphId).After(oldTimestamp) {
				g.StoreBucket(src, stage, oldGraphBucket)
				oldGraphBucket.Set(newGraphId, oldTimestamp)

				if newGraphBucket != nil {
					newGraphBucket.Delete(newGraphId)
				}
			}

			oldGraphBucket.Delete(oldGraphID)
		}

		for stage, oldGraphBucket := range otherGraph.GetBuckets(dst) {
			newGraphBucket := g.GetBucket(dst, stage)
			oldTimestamp := oldGraphBucket.Get(newGraphId)

			if newGraphBucket == nil || newGraphBucket.Get(newGraphId).After(oldTimestamp) {
				g.StoreBucket(dst, stage, oldGraphBucket)
				oldGraphBucket.Set(newGraphId, oldTimestamp)

				if newGraphBucket != nil {
					newGraphBucket.Delete(newGraphId)
				}
			}

			oldGraphBucket.Delete(oldGraphID)
		}
	}

	for stage, otherGraphStageRelevance := range otherGraph.relevances {
		graphStageRelevance := g.relevances[stage]

		if graphStageRelevance < otherGraphStageRelevance {
			g.relevances[stage] = otherGraphStageRelevance
			g.ComputedRelevance += (otherGraphStageRelevance - graphStageRelevance) * stage.GetWeight()
		}
	}
}

type PreComputedDirectedRelation[T Stage] struct {
	ID string `json:"id"`
	DirectedRelationJson[T]
	ComputedHostRelevance float32    `json:"computed_host_relevance"`
	ConfirmedStages       []UKCStage `json:"confirmed_ukc_stages"`
	FromIsInternal        bool       `json:"from_is_internal"`
	ToIsInternal          bool       `json:"to_is_internal"`
	FromRiskLevel         RiskLevel  `json:"from_risk_level"`
	ToRiskLevel           RiskLevel  `json:"to_risk_level"`
}

func (r *OptimizedDirectedRelation[T]) getConfirmedStages(hasLateralMovement bool, hasOutgoingActivity bool) []UKCStage {
	stages := r.MetaStage.ToUKCStages()

	confirmedStages := []UKCStage{}
	for _, stage := range stages {
		switch stage {
		case O:
			if !hasOutgoingActivity {
				confirmedStages = append(confirmedStages, stage)
			}
		case E:
			if !hasLateralMovement {
				confirmedStages = append(confirmedStages, stage)
			}
		default:
			confirmedStages = append(confirmedStages, stage)
		}
	}

	return confirmedStages
}

func (r *OptimizedDirectedRelation[T]) GetPreComputed(stages []UKCStage, count int, id OptimizedDirectedRelationID) PreComputedDirectedRelation[T] {
	src := id.GetSrc()
	dst := id.GetDst()
	relevance := id.Relevance()

	victim := id.GetSrc()
	if r.MetaStage.GetVictim() == Destination {
		victim = dst
	}

	return PreComputedDirectedRelation[T]{
		ID: base64.StdEncoding.EncodeToString(id[:]),
		DirectedRelationJson: DirectedRelationJson[T]{
			From:        src.String(),
			To:          dst.String(),
			MetaStage:   r.MetaStage,
			Timestamp:   r.Timestamp.UnixMilli(),
			Severity:    id.Severity(),
			Confidence:  id.Confidence(),
			SignatureId: id.GetSignatureId(),
			Cause:       r.Cause,
			Labels:      r.GetLabels(),
			Count:       count,
		},
		ComputedHostRelevance: relevance * float32(HostManager.GetHostRiskLevel(victim)),
		ConfirmedStages:       stages,
		FromIsInternal:        src.IsInternal(),
		ToIsInternal:          dst.IsInternal(),
		FromRiskLevel:         HostManager.GetHostRiskLevel(src),
		ToRiskLevel:           HostManager.GetHostRiskLevel(dst),
	}
}

type PreComputedGraph[T Stage] struct {
	PreComputedDirectedRelations []PreComputedDirectedRelation[T] `json:"relations"`
	ComputedRelevance            float32                          `json:"computed_relevance"`
}

func (g *PreComputedGraph[T]) Len() int {
	return len(g.PreComputedDirectedRelations)
}

func (g *PreComputedGraph[T]) Swap(i int, j int) {
	a := g.PreComputedDirectedRelations[i]
	g.PreComputedDirectedRelations[i] = g.PreComputedDirectedRelations[j]
	g.PreComputedDirectedRelations[j] = a
}

func (g *PreComputedGraph[T]) Less(i, j int) bool {
	return g.PreComputedDirectedRelations[i].Timestamp < g.PreComputedDirectedRelations[j].Timestamp
}

func (g *Graph[T, K]) GetPreComputed() PreComputedGraph[T] {
	graph := PreComputedGraph[T]{
		PreComputedDirectedRelations: []PreComputedDirectedRelation[T]{},
		ComputedRelevance:            g.ComputedRelevance,
	}

	g.relationsMutex.RLock()
	for id, relation := range g.Relations {
		count := relation.Count

		srcEntry := g.reverseLookup[id.GetSrc()]
		dstEntry := g.reverseLookup[id.GetDst()]

		stages := relation.getConfirmedStages(srcEntry.HasLateralMovement, dstEntry.HasOutgoingActivity)

		graph.PreComputedDirectedRelations = append(graph.PreComputedDirectedRelations, relation.GetPreComputed(stages, count, id))
	}
	g.relationsMutex.RUnlock()

	sort.Sort(&graph)

	return graph
}

type GraphJson[T Stage] struct {
	Relations         []DirectedRelationJson[T] `json:"relations"`
	ComputedRelevance float32                   `json:"computed_relevance"`
}

func (g *Graph[T, K]) MarshalJSON() ([]byte, error) {
	g.relationsMutex.RLock()
	defer g.relationsMutex.RUnlock()

	graph := GraphJson[T]{}

	for id, relation := range g.Relations {
		graph.Relations = append(graph.Relations, DirectedRelationJson[T]{
			From:        id.GetSrc().String(),
			To:          id.GetDst().String(),
			Timestamp:   relation.Timestamp.UnixMilli(),
			MetaStage:   relation.MetaStage,
			Severity:    id.Severity(),
			Confidence:  id.Confidence(),
			SignatureId: id.GetSignatureId(),
			Cause:       relation.Cause,
			Labels:      relation.GetLabels(),
			Count:       relation.Count,
		})
	}

	graph.ComputedRelevance = g.ComputedRelevance

	return json.Marshal(graph)
}

func (g *Graph[T, K]) UnmarshalJSON(data []byte) error {
	jsonObject := GraphJson[T]{}
	err := json.Unmarshal(data, &jsonObject)
	if err != nil {
		return err
	}

	g.relevances = map[T]float32{}
	g.reverseLookup = map[IPAddress]ReverseLookupEntry[K]{}
	g.relationsMutex = &sync.RWMutex{}
	g.Relations = map[OptimizedDirectedRelationID]OptimizedDirectedRelation[T]{}

	for _, relation := range jsonObject.Relations {
		g.append(DirectedRelation[T]{
			SrcNode:     ParseIPAddress(relation.From),
			DstNode:     ParseIPAddress(relation.To),
			Timestamp:   time.UnixMilli(relation.Timestamp),
			MetaStage:   relation.MetaStage,
			Severity:    relation.Severity,
			Confidence:  relation.Confidence,
			SignatureId: relation.SignatureId,
			Cause:       relation.Cause,
			Labels:      relation.Labels,
		})
	}

	return nil
}
