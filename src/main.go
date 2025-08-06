package main

import (
	"embed"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
	"rtkcsm/connector/transport"
	"rtkcsm/connector/visualization"
	"runtime/pprof"
	"slices"
	"time"

	"github.com/alexflint/go-arg"
)

//go:embed static
var assets embed.FS

type configuration struct {
	TransportFilePath          string             `arg:"--file" help:"Log files from suricata (eve.json) or zeek (JSON format)"`
	TransportListenAddress     string             `arg:"--listen" help:"TCP Port to listen on for ingesting from Tenzir"`
	VisualizationListenAddress string             `arg:"--server" help:"Port for web interface for visualization"`
	ImportGraphsFile           string             `arg:"--import" help:"Import existing graphs from RT-KCSM's 'graphs.json'"`
	ReaderType                 string             `arg:"--reader" help:"Format for reading from Tenzir/File: 'zeek', 'suricata', 'ocsf', 'suricata-tenzir'"`
	TransportType              string             `arg:"--transport" help:"Specify whether to use file or TCP for ingestion"`
	ExportGraphsFile           string             `arg:"--export" help:"File name of exported graphs from RT-KCSM"`
	HostRisk                   map[string]float32 `arg:"--risk" help:"Define risk score (low=0.5,default=1.0,high=1.5) of an IP address for a host/asset."`
	ProfilerOptions            map[string]string  `arg:"--profile" help:"performance profile options: memory=/path/to/file, cpu=/path/to/file, alerts=/path/to/file, graphs=/path/to/file, graph-ranking=/path/to/file, progress=true"`
	ProfilerGraphID            structure.GraphID  `arg:"--profile-graph-ranking-id" help:"graph id for profiling ranking"`
	StageWeights               map[string]float32 `arg:"--stage-weight" help:"define custom stage weights (incoming, same-zone, different-zone, outgoing)"`
	ProfilerLogResolution      int                `arg:"--profile-log-resolution" help:"Resolution of updating alert count" default:"1000"`
}

func startCPUProfile(fileName string) *os.File {
	file, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		log.Panicf("could not create cpu profile: %s", err)
		return nil
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		log.Panicf("could not start cpu profile: %s", err)
		return nil
	}

	return file
}

func stopCPUProfile(file *os.File) {
	pprof.StopCPUProfile()
	if file != nil {
		if err := file.Close(); err != nil {
			log.Panic(err)
		}
		fmt.Printf("cpu profile written to: %s", file.Name())
	}
}

func takeSnapshortOfHeapProfile(fileName string) {
	file, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		log.Panicf("could not create memory heap profile: %s", err)
	}

	if err := pprof.WriteHeapProfile(file); err != nil {
		log.Panicf("could not write memory heap profile: %s", err)
	}
}

func main() {
	var config configuration
	arg.MustParse(&config)

	profilerOptions := structure.ParseProfilerOptions(slices.Collect(maps.Keys(config.ProfilerOptions)), config.ProfilerGraphID, config.ProfilerLogResolution)

	if profilerOptions.HasAny() {
		log.Printf("Profiling is on (this can affect performance): %s\n", profilerOptions.String())
	}

	if profilerOptions.Has(structure.CpuProfilerOptionFlag) {
		profileFile := startCPUProfile(config.ProfilerOptions[string(structure.CpuOption)])
		defer stopCPUProfile(profileFile)
	}

	if profilerOptions.Has(structure.MemoryProfilerOptionFlag) {
		file, err := os.Create(filepath.Clean(config.ProfilerOptions[string(structure.MemoryOption)]))
		if err != nil {
			log.Panic(err)
		}

		err = structure.PerformanceManager.GetSeries("memory").Start(file)
		if err != nil {
			log.Panic(err)
		}
	}

	if profilerOptions.Has(structure.AlertsProfilerOptionFlag) {
		file, err := os.Create(filepath.Clean(config.ProfilerOptions[string(structure.AlertsOption)]))
		if err != nil {
			log.Panic(err)
		}

		err = structure.PerformanceManager.GetSeries("alerts").Start(file)
		if err != nil {
			log.Panic(err)
		}
	}

	if profilerOptions.Has(structure.GraphRankingProfilerOptionFlag) {
		file, err := os.Create(filepath.Clean(config.ProfilerOptions[string(structure.GraphRankingOption)]))
		if err != nil {
			log.Panic(err)
		}

		err = structure.PerformanceManager.GetSeries("graph-ranking").Start(file)
		if err != nil {
			log.Panic(err)
		}
	}

	if profilerOptions.Has(structure.GraphNumberProfilerOptionFlag) {
		file, err := os.Create(filepath.Clean(config.ProfilerOptions[string(structure.GraphNumberOption)]))
		if err != nil {
			log.Panic(err)
		}
		err = structure.PerformanceManager.GetSeries("graph-count").Start(file)
		if err != nil {
			log.Panic(err)
		}
	}

	for ipAdress, risk := range config.HostRisk {
		structure.HostManager.AddHostRiskLevel(structure.ParseIPAddress(ipAdress), structure.RiskLevel(risk))
	}

	for stage, weight := range config.StageWeights {
		structure.SimplifiedUkcStageWeights[structure.NewSimplifiedUKCStageFromString(stage)] = weight
	}

	rtkcsm := behaviour.NewIncrementalRTKCSM(128, structure.NewSimplifiedUKCStageMapper(), structure.NewUKCStateMachine[structure.SimplifiedUKCStage](), &profilerOptions)

	startTime := time.Now()
	if config.ImportGraphsFile != "" {
		file, err := os.Open(config.ImportGraphsFile)
		if err != nil {
			log.Printf("could not load save: %s", err)
		} else {
			err = rtkcsm.ImportGraphs(file)
			if err != nil {
				log.Printf("could not load save: %s", err)
			} else {
				log.Println("import done")
			}

			if err := file.Close(); err != nil {
				log.Panic(err)
			}
		}

	}

	var alertReader reader.AlertReader[structure.SimplifiedUKCStage, structure.UKCStage]
	switch config.ReaderType {
	case "", "suricata":
		alertReader = &reader.SuricataAlertReader[structure.SimplifiedUKCStage, structure.UKCStage]{}
	case "zeek":
		alertReader = &reader.ZeekAlertReader[structure.SimplifiedUKCStage, structure.UKCStage]{}
	case "suricata-tenzir":
		alertReader = &reader.SuricataTenzirAlertReader[structure.SimplifiedUKCStage, structure.UKCStage]{}
	case "ocsf":
		alertReader = &reader.OCSFAlertReader[structure.SimplifiedUKCStage, structure.UKCStage]{}
	default:
		log.Panic("reader type is not known")
	}

	if config.VisualizationListenAddress != "" {
		go visualization.Start(config.VisualizationListenAddress, rtkcsm, assets)
	}

	var selectedTransport transport.Transport[structure.SimplifiedUKCStage, structure.UKCStage]
	switch config.TransportType {
	case "", "file":
		if config.TransportFilePath != "" {
			selectedTransport = &transport.FileTransport[structure.SimplifiedUKCStage, structure.UKCStage]{
				FilePath: config.TransportFilePath,
			}
		}
	case "tcp":
		selectedTransport = &transport.TcpTransport[structure.SimplifiedUKCStage, structure.UKCStage]{
			ListenAddress: config.TransportListenAddress,
		}
	default:
		log.Panic("transport type is not known")
	}

	if selectedTransport != nil {
		err := selectedTransport.Start(rtkcsm, alertReader)
		if err != nil {
			log.Panicf("error channeling alerts: %s", err)
		}
	}

	endTime := time.Now()
	graphCount := rtkcsm.GetGraphList(-1).Count
	err := profilerOptions.TakeMeasurement(graphCount, true)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Generation is done: Execution Time: %.02fs, Generated graphs: %d", endTime.Sub(startTime).Seconds(), graphCount)

	structure.PerformanceManager.StopAllSeries()

	if profilerOptions.Has(structure.MemoryAllocationProfilerOptionFlag) {
		takeSnapshortOfHeapProfile(string(structure.MemoryAllocationOption))
	}

	if config.ExportGraphsFile != "" {
		filePath := filepath.Clean(config.ExportGraphsFile)
		file, err := os.Create(filePath)
		if err != nil {
			log.Panic(err)
		}

		size, err := rtkcsm.ExportGraphs(file)
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Exported graphs file to %s (%.02f MB)", filePath, float32(size)/1024/1024)
	}

	if config.VisualizationListenAddress != "" {
		<-make(chan struct{})
	}
}
