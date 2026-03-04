package behaviour

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"rtkcsm/component/structure"
	"strconv"
	"strings"
	"sync"
)

type RTKCSMImplementation[T structure.Stage, K structure.Stage] struct {
	sortedGraphs    structure.SortedMap[structure.GraphID, float32]
	graphs          map[structure.GraphID]*structure.Graph[T, K]
	lookup          structure.LookupTable[T, K]
	graphsMutex     *sync.RWMutex
	profilerOptions *structure.ProfilerOptions
	stageMapper     structure.StageMapper[T]
	stateMachine    structure.StateMachine[T, K]
}

func NewIncrementalRTKCSM[T structure.Stage, K structure.Stage](workerCount int, stageMapper structure.StageMapper[T], stateMachine structure.StateMachine[T, K], profilerOptions *structure.ProfilerOptions) *RTKCSMImplementation[T, K] {
	var sortedGraphs structure.SortedMap[structure.GraphID, float32]
	sortedGraphs = structure.NewWriteEfficientSortedMap[structure.GraphID, float32](true)
	if profilerOptions.Has(structure.GraphRankingProfilerOptionFlag) {
		sortedGraphs = structure.NewReadEfficientSortedMap[structure.GraphID, float32](true)
	}

	rtkcsm := RTKCSMImplementation[T, K]{
		graphs:          map[structure.GraphID]*structure.Graph[T, K]{},
		lookup:          structure.NewLookupTable(stateMachine),
		graphsMutex:     &sync.RWMutex{},
		sortedGraphs:    sortedGraphs,
		profilerOptions: profilerOptions,
		stageMapper:     stageMapper,
	}

	return &rtkcsm
}

func (c *RTKCSMImplementation[T, K]) GetGraphList(page int) structure.GraphInformationList {
	graphs := []structure.GraphInformation{}
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()

	length := c.sortedGraphs.Len()

	if page < 0 {
		page = 0
	}

	itemsPerPage := 100
	offset := page * itemsPerPage
	limit := offset + itemsPerPage

	if limit >= length {
		limit = length
	}

	for i := offset; i < limit; i++ {
		id, relevance := c.sortedGraphs.Get(i)

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

func (c *RTKCSMImplementation[T, K]) GetGraph(id structure.GraphID) *structure.Graph[T, K] {
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()
	graph, ok := c.graphs[id]

	if ok {
		return graph
	}

	return nil
}

func (c *RTKCSMImplementation[T, K]) Reset() {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	c.graphs = map[structure.GraphID]*structure.Graph[T, K]{}
	c.lookup = structure.NewLookupTable(c.stateMachine)
	c.sortedGraphs = structure.NewWriteEfficientSortedMap[structure.GraphID, float32](true)
}

func (c *RTKCSMImplementation[T, K]) AddAlert(alert structure.Alert) error {
	err := c.processStages(alert)
	if err != nil {
		return err
	}
	return c.profilerOptions.TakeMeasurement(len(c.graphs), false)
}

func (c *RTKCSMImplementation[T, K]) processStages(alert structure.Alert) error {
	if !alert.SourceIP.IsUnspecified() && !alert.DestinationIP.IsUnspecified() {
		metaStage, err := c.stageMapper.DetermineStage(alert)
		if err == nil {
			relation := structure.EnrichedAlert[T]{
				Alert:     alert,
				MetaStage: metaStage,
			}

			c.processRelation(relation)

			return nil
		}

		return fmt.Errorf("stage not found (%s) -> (%s): %s", alert.SourceIP, alert.DestinationIP, err)
	} else {
		return fmt.Errorf("undefined source (%s) or destination ip (%s)", alert.SourceIP, alert.DestinationIP)
	}
}

func (c *RTKCSMImplementation[T, K]) processRelation(alert structure.EnrichedAlert[T]) {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	graphIds := c.lookup.SearchRelations(&alert).ToSlice()

	var graphId structure.GraphID = 0
	var graph *structure.Graph[T, K]

	if len(graphIds) == 0 {
		graphId = structure.NextGraphID()
		graph = structure.NewGraph[T, K]()
		c.graphs[graphId] = graph
	} else if len(graphIds) == 1 {
		graphId = graphIds[0]
		graph = c.graphs[graphId]
	} else {
		// merge graphs if there is an overlap
		for _, duplicateGraphId := range graphIds {
			if graphId > duplicateGraphId || graphId == 0 {
				graphId = duplicateGraphId
			}
		}

		graph = c.graphs[graphId]

		for _, duplicateGraphId := range graphIds {
			if duplicateGraphId != graphId {
				graph.Merge(c.graphs[duplicateGraphId], duplicateGraphId, graphId)
				c.sortedGraphs.Delete(duplicateGraphId)
				delete(c.graphs, duplicateGraphId)
			}
		}

	}
	relation := graph.Append(alert)

	// update sorted graphs
	c.sortedGraphs.Insert(graphId, graph.Relevance())

	c.lookup.AddRelation(&relation, graphId, graph)

	// Get position of correct graph for eval
	if c.profilerOptions.Has(structure.GraphRankingProfilerOptionFlag) {
		series := structure.PerformanceManager.GetSeries("graph-ranking")
		if series != nil {
			series.Add(relation.Timestamp, c.sortedGraphs.GetPosition(c.profilerOptions.GetGraphId()), len(c.graphs))
		}
	}
}

func (c *RTKCSMImplementation[T, K]) ImportGraphs(reader io.Reader) error {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	scanner := bufio.NewScanner(reader)
	bufferSize := 2097152 // 2 MB buffer
	buffer := make([]byte, bufferSize)
	scanner.Buffer(buffer, bufferSize)

	for scanner.Scan() {
		line := scanner.Text()

		splits := strings.SplitN(line, ",", 2)
		if len(splits) != 2 {
			return fmt.Errorf("wrong format of line: splits %d != 2", len(splits))
		}

		id, err := strconv.Atoi(splits[0])
		if err != nil {
			return fmt.Errorf("error converting graph id: %s", err)
		}

		var graph structure.Graph[T, K]
		err = json.Unmarshal([]byte(splits[1]), &graph)
		if err != nil {
			return err
		}

		graphId := structure.GraphID(id)
		c.graphs[graphId] = &graph
		for _, relation := range graph.GetRelations() {
			c.lookup.AddRelation(&relation, graphId, &graph)
		}

		c.sortedGraphs.Insert(graphId, graph.Relevance())
	}

	return scanner.Err()
}

func (c *RTKCSMImplementation[T, K]) ExportGraphs(writer io.Writer) (int, error) {
	c.graphsMutex.RLock()
	defer c.graphsMutex.RUnlock()

	totalSize := 0

	for id, graph := range c.graphs {
		text, err := json.Marshal(graph)
		if err != nil {
			return totalSize, fmt.Errorf("error encoding JSON: %s", err)
		}

		size, err := writer.Write(append([]byte(strconv.Itoa(int(id))+","), append(text, []byte("\n")...)...))
		if err != nil {
			return totalSize, fmt.Errorf("error writing JSON: %s", err)
		}

		totalSize += size
	}

	return totalSize, nil
}

func (c *RTKCSMImplementation[T, K]) GetHostRisks() []structure.HostRisk {
	hostRisks := []structure.HostRisk{}

	for host, risk := range structure.HostManager.GetHosts() {
		hostRisks = append(hostRisks, structure.HostRisk{
			IpAddress: host.String(),
			RiskLevel: float32(risk),
		})
	}

	return hostRisks
}

func (c *RTKCSMImplementation[T, K]) recomputeGraphRelevances() {
	c.graphsMutex.Lock()
	defer c.graphsMutex.Unlock()
	for graphID, graph := range c.graphs {
		c.sortedGraphs.Insert(graphID, graph.RecomputeRelevance())
	}
}

func (c *RTKCSMImplementation[T, K]) AddHostRisk(address structure.IPAddress, riskLevel structure.RiskLevel) {
	structure.HostManager.AddHostRiskLevel(address, riskLevel)
	c.recomputeGraphRelevances()
}

func (c *RTKCSMImplementation[T, K]) DeleteHostRisk(address structure.IPAddress) {
	structure.HostManager.DeleteHostRiskLevel(address)
	c.recomputeGraphRelevances()
}
