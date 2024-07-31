package behaviour

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"rtkcsm/component/structure"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

var graphsPerPage = 100

type IncrementalRTKCSM struct {
	sortedGraphs     structure.SortedMap[structure.GraphID, float32]
	graphs           map[structure.GraphID]*structure.Graph
	lookup           structure.LookupTable
	stagesWorkerPool *structure.WorkerPool[structure.Alert]
	graphWorkerPool  *structure.WorkerPool[structure.DirectedRelation]
	graphsMutex      *sync.RWMutex
	profilerOptions  structure.ProfilerOptions
	eventManager     *structure.EventManager
}

func NewIncrementalRTKCSM(workerCount int, profilerOptions structure.ProfilerOptions) *IncrementalRTKCSM {
	rtkcsm := IncrementalRTKCSM{
		graphs:          map[structure.GraphID]*structure.Graph{},
		lookup:          structure.NewLookupTable(),
		graphsMutex:     &sync.RWMutex{},
		sortedGraphs:    structure.NewWriteEfficientSortedMap[structure.GraphID, float32](true),
		profilerOptions: profilerOptions,
		eventManager:    structure.NewEventManager(),
	}

	rtkcsm.stagesWorkerPool = structure.NewWorkerPool(workerCount, rtkcsm.processStages, profilerOptions)
	rtkcsm.graphWorkerPool = structure.NewWorkerPool(workerCount, rtkcsm.processGraphs, profilerOptions)

	return &rtkcsm
}

func (c *IncrementalRTKCSM) GetGraphList(page int) structure.GraphInformationList {
	graphs := []structure.GraphInformation{}
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()

	length := c.sortedGraphs.Len()
	offset := page * graphsPerPage
	limit := offset + graphsPerPage

	if page < 0 || limit > length {
		offset = 0
		limit = length
	}

	for index := range limit - offset {
		id, relevance := c.sortedGraphs.Get(index + offset)

		graphs = append(graphs, structure.GraphInformation{
			ID:        id,
			Relevance: relevance,
		})
	}

	return structure.GraphInformationList{
		Graphs: graphs,
		Count:  length,
	}
}

func (c *IncrementalRTKCSM) GetGraph(id structure.GraphID) *structure.Graph {
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()
	graph, ok := c.graphs[id]

	if ok {
		return graph
	}

	return nil
}

func (c *IncrementalRTKCSM) Reset() {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	c.graphs = map[structure.GraphID]*structure.Graph{}
	c.lookup = structure.NewLookupTable()
	c.sortedGraphs = structure.NewWriteEfficientSortedMap[structure.GraphID, float32](true)
}

func (c *IncrementalRTKCSM) exportJSON() error {
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()
	file, err := os.Create("graphs.json")
	if err != nil {
		return fmt.Errorf("error creating file for export: %s", err)
	}
	defer file.Close()

	indentedJSON, err := json.MarshalIndent(c.graphs, "", "    ")
	if err != nil {
		return fmt.Errorf("error encoding JSON: %s", err)
	}

	res, err := file.Write(indentedJSON)
	if err != nil {
		return fmt.Errorf("error writing JSON: %s", err)
	}
	fmt.Printf("successfully wrote %d bytes\n", res)

	return nil
}

func (c *IncrementalRTKCSM) exportDot() error {
	err := os.MkdirAll("graphs", 0750)
	if err != nil {
		return err
	}
	count := 0
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()
	for id, t := range c.graphs {
		fileName := filepath.Clean(filepath.Join("graphs", fmt.Sprintf("%d-%d", count, id)+".dot"))
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(t.ExportDot(false))
		if err != nil {
			return err
		}
		count++
	}
	return nil
}

func (c *IncrementalRTKCSM) Add(alert structure.Alert) {
	c.stagesWorkerPool.AddData(alert)
}

func (c *IncrementalRTKCSM) processStages(alert structure.Alert) {
	if !alert.SourceIP.IsUnspecified() && !alert.DestinationIP.IsUnspecified() {
		stage := structure.DetermineStage(alert.SourceIP, alert.DestinationIP)

		var id string

		if c.profilerOptions.Has(structure.DoNotOptimizeGraphsProfilerOptionFlag) {
			id = uuid.NewString()
		} else {
			id = fmt.Sprintf("%s-%s-%d-%f", alert.SourceIP.String(), alert.DestinationIP.String(), stage, alert.Severity)
		}

		relation := structure.DirectedRelation{
			ID:           structure.DirectedRelationID(id),
			DstNode:      alert.DestinationIP,
			SrcNode:      alert.SourceIP,
			Timestamp:    alert.Timestamp,
			Stage:        stage,
			Severity:     alert.Severity,
			TruePositive: alert.TruePositive,
		}

		if relation.Stage != structure.None {
			c.graphWorkerPool.AddData(relation)
		}
	}
}

func (c *IncrementalRTKCSM) processGraphs(relation structure.DirectedRelation) {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	graphIds := c.lookup.GetGraphIds(&relation)

	var graphID structure.GraphID = 0
	var graph *structure.Graph

	switch graphIds.Size() {
	case 0:
		graphID = structure.NextGraphID()
		graph = structure.NewGraph(relation)
		c.graphs[graphID] = graph
	case 1:
		graphID = *graphIds.First()
		graph = c.graphs[graphID]
		graph.Append(relation)
		c.graphs[graphID] = graph
	default:
		// merge graphs if there is an overlap
		graph = structure.NewGraph(relation)

		for duplicateGraphId := range graphIds {
			graph.Merge(c.graphs[duplicateGraphId])

			if graphID > duplicateGraphId || graphID == 0 {
				if graphID > 0 {
					delete(c.graphs, graphID)
					c.sortedGraphs.Delete(graphID)
				}

				graphID = duplicateGraphId
			} else {
				delete(c.graphs, duplicateGraphId)
				c.sortedGraphs.Delete(duplicateGraphId)
			}
		}

		c.graphs[graphID] = graph
		c.lookup.MergeGraphs(graphIds, graphID)
	}

	relevance := graph.Relevance()

	// update sorted graphs
	c.sortedGraphs.Insert(graphID, relevance)

	c.lookup.AddRelation(&relation, graphID)
	c.eventManager.Publish(graphID, structure.NewDirectedRelationEvent{
		NewDirectedRelationEventData: structure.NewDirectedRelationEventData{
			DirectedRelation: relation,
			GraphRelevance:   relevance,
		},
		GraphID: graphID,
	})

	// Get position of correct graph for eval
	if c.profilerOptions.Has(structure.GraphRankingProfilerOptionFlag) {
		structure.PerformanceManager.GetSeries("graph-ranking").Add(relation.Timestamp, c.sortedGraphs.GetPosition(c.profilerOptions.GetGraphId()), len(c.graphs))
	}

	// Get graph count
	if c.profilerOptions.Has(structure.GraphNumberProfilerOptionFlag) {
		structure.PerformanceManager.GetSeries("graph-count").Add(time.Now(), len(c.graphs))
	}
}

func (c *IncrementalRTKCSM) ImportSave(reader io.Reader) error {
	graphs := map[structure.GraphID]structure.Graph{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&graphs); err != nil {
		log.Printf("could not read graphs because of %s", err)
	}

	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()

	for graphID, graph := range graphs {
		graph := &graph
		c.graphs[graphID] = graph
		for _, relation := range graph.GetRelations() {
			c.lookup.AddRelation(&relation, graphID)
		}

		c.sortedGraphs.Insert(graphID, graph.Relevance())
	}

	return nil
}

func (c *IncrementalRTKCSM) Stop(exportFormats []structure.ExportFormat) {
	c.stagesWorkerPool.Stop()
	c.graphWorkerPool.Stop()

	if slices.Contains(exportFormats, structure.ExportFormatJson) {
		err := c.exportJSON()
		if err != nil {
			fmt.Println(err)
		}
	}

	if slices.Contains(exportFormats, structure.ExportFormatDot) {
		err := c.exportDot()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c *IncrementalRTKCSM) GetHostRisks() []structure.HostRisk {
	hostRisks := []structure.HostRisk{}

	for host, risk := range structure.HostManager.GetHosts() {
		hostRisks = append(hostRisks, structure.HostRisk{
			IpAddress: host.String(),
			RiskLevel: float32(risk),
		})
	}

	return hostRisks
}

func (c *IncrementalRTKCSM) recomputeGraphRelevances() {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	for graphID, graph := range c.graphs {
		c.sortedGraphs.Insert(graphID, graph.RecomputeRelevance())
	}
}

func (c *IncrementalRTKCSM) AddHostRisk(address structure.IPAddress, riskLevel structure.RiskLevel) {
	structure.HostManager.AddHostRiskLevel(address, riskLevel)
	c.recomputeGraphRelevances()
}

func (c *IncrementalRTKCSM) DeleteHostRisk(address structure.IPAddress) {
	structure.HostManager.DeleteHostRiskLevel(address)
	c.recomputeGraphRelevances()
}

func (c *IncrementalRTKCSM) GetEventManager() *structure.EventManager {
	return c.eventManager
}
