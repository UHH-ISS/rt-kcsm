package structure

import (
	"encoding/json"
	"strings"
	"sync"
)

type GraphID int

var stageWeights = map[SimplifiedUKCStage]float32{
	Recon:        0.5,
	Host:         1,
	Lateral:      1.25,
	Pivot:        1.5,
	Exfiltration: 2,
}

var nextGraphID GraphID = 0

func NextGraphID() GraphID {
	nextGraphID += 1
	return nextGraphID
}

type GraphInformationList struct {
	Graphs []GraphInformation `json:"graphs"`
	Count  int                `json:"count"`
}

type GraphInformation struct {
	ID        GraphID `json:"id"`
	Relevance float32 `json:"relevance"`
}

type Graph struct {
	Relations         map[DirectedRelationID]DirectedRelation `json:"relations"`
	relevances        map[SimplifiedUKCStage]float32
	relationsMutex    *sync.RWMutex
	ComputedRelevance float32 `json:"computed_relevance"`
}

func NewGraph(relation DirectedRelation) *Graph {
	graph := Graph{
		Relations:      map[DirectedRelationID]DirectedRelation{},
		relevances:     map[SimplifiedUKCStage]float32{},
		relationsMutex: &sync.RWMutex{},
	}

	graph.Append(relation)

	return &graph
}

func (g *Graph) Relevance() float32 {
	return g.ComputedRelevance
}

func (g *Graph) GetRelations() map[DirectedRelationID]DirectedRelation {
	return g.Relations
}

func (g *Graph) Append(relation DirectedRelation) {
	g.relationsMutex.Lock()
	defer g.relationsMutex.Unlock()

	g.Relations[relation.ID] = relation
	relationRelevance := relation.Relevance()
	existingMaxStageRelevance := g.relevances[relation.Stage]

	if existingMaxStageRelevance < relationRelevance {
		g.relevances[relation.Stage] = relationRelevance
		g.ComputedRelevance += (relationRelevance - existingMaxStageRelevance) * stageWeights[relation.Stage]
	}
}

func (g *Graph) RecomputeRelevance() float32 {
	g.relationsMutex.Lock()
	defer g.relationsMutex.Unlock()
	g.relevances = map[SimplifiedUKCStage]float32{}
	g.ComputedRelevance = 0

	for _, relation := range g.Relations {
		relationRelevance := relation.Relevance()
		existingMaxStageRelevance := g.relevances[relation.Stage]

		if existingMaxStageRelevance < relationRelevance {
			g.relevances[relation.Stage] = relationRelevance
			g.ComputedRelevance += (relationRelevance - existingMaxStageRelevance) * stageWeights[relation.Stage]
		}
	}

	return g.ComputedRelevance
}

func (g *Graph) Merge(otherGraph *Graph) {
	g.relationsMutex.Lock()
	defer g.relationsMutex.Unlock()
	for _, relation := range otherGraph.Relations {
		g.Relations[relation.ID] = relation
	}

	for stage, otherGraphStageRelevance := range otherGraph.relevances {
		graphStageRelevance := g.relevances[stage]

		if graphStageRelevance < otherGraphStageRelevance {
			g.relevances[stage] = otherGraphStageRelevance
			g.ComputedRelevance += (otherGraphStageRelevance - graphStageRelevance) * stageWeights[stage]
		}
	}
}

func (g *Graph) ExportDot(simplify bool) string {
	dotGraph := []string{"digraph {", "node [shape=circle fontsize=16]", "edge [length=100, color=gray, fontcolor=black]"}

	g.relationsMutex.RLock()
	for _, relation := range g.Relations {
		dotGraph = append(dotGraph, relation.toDot(simplify))
	}
	g.relationsMutex.RUnlock()

	dotGraph = append(dotGraph, "}")

	return strings.Join(dotGraph, "\n")
}

type GraphJson struct {
	Relations         []DirectedRelation `json:"relations"`
	ComputedRelevance float32            `json:"computed_relevance"`
}

func (g *Graph) MarshalJSON() ([]byte, error) {
	graph := GraphJson{}

	for _, relation := range g.Relations {
		graph.Relations = append(graph.Relations, relation)
	}

	graph.ComputedRelevance = g.ComputedRelevance

	return json.Marshal(graph)
}

func (g *Graph) UnmarshalJSON(data []byte) error {
	jsonObject := GraphJson{}
	err := json.Unmarshal(data, &jsonObject)
	if err != nil {
		return err
	}

	g.relevances = map[SimplifiedUKCStage]float32{}
	g.Relations = map[DirectedRelationID]DirectedRelation{}
	g.relationsMutex = &sync.RWMutex{}

	for _, relation := range jsonObject.Relations {
		g.Append(relation)
	}

	return nil
}
