package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"rtkcsm/component/behaviour"
	"rtkcsm/component/structure"
	"rtkcsm/connector/reader"
	"rtkcsm/connector/transport"
	"rtkcsm/connector/visualization"
	"runtime/pprof"
	"time"

	"github.com/alexflint/go-arg"
)

//go:embed static
var assets embed.FS

type configuration struct {
	TransportFilePath          string                   `arg:"--file" help:"Log files from suricata (eve.json) or zeek (JSON format)"`
	TransportListenAddress     string                   `arg:"--listen" help:"TCP Port to listen on for ingesting from Tenzir"`
	VisualizationListenAddress string                   `arg:"--server" help:"Port for web interface for visualization"`
	GraphsFile                 string                   `arg:"--import" help:"Import existing graphs from RT-KCSM's 'graphs.json'"`
	ReaderType                 string                   `arg:"--reader" help:"Format for reading from Tenzir/File: 'zeek', 'suricata', 'ocsf', 'suricata-tenzir'"`
	TransportType              string                   `arg:"--transport" help:"Specify whether to use file or TCP for ingestion"`
	ExportFormats              []structure.ExportFormat `arg:"--export" help:"Export graphs from RT-KCSM"`
	ValuableAssets             []string                 `arg:"--valuable-assets" help:"Define IP address of valuable asset that should be assigned with a high risk score"`
	ProfilerOptions            []string                 `arg:"--profile" help:"performance profile options: memory, cpu, alerts, graph-ranking, no-graph-optimization"`
	ProfilerGraphID            structure.GraphID        `arg:"--profile-graph-ranking-id" help:"graph id for profiling ranking"`
}

func startCPUProfile() *os.File {
	file, err := os.Create(fmt.Sprintf("cpu_profile_%s.pprof", time.Now()))
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
		fmt.Printf("CPU profile written to: %s", file.Name())
	}
}

func main() {
	startTime := time.Now()
	var config configuration
	arg.MustParse(&config)

	profilerOptions := structure.ParseProfilerOptions(config.ProfilerOptions, config.ProfilerGraphID)

	if profilerOptions.HasAny() {
		log.Printf("Profiling is on (this might affect performance): %s\n", profilerOptions.String())
	}

	if profilerOptions.Has(structure.CpuProfilerOptionFlag) {
		profileFile := startCPUProfile()
		defer stopCPUProfile(profileFile)
	}

	for _, ipAdress := range config.ValuableAssets {
		structure.HostManager.AddHostRiskLevel(structure.ParseIPAddress(ipAdress), structure.HighRisk)
	}

	rtkcsm := behaviour.NewIncrementalRTKCSM(1, profilerOptions)

	if config.GraphsFile != "" {
		file, err := os.Open(config.GraphsFile)
		if err != nil {
			log.Printf("could not load save: %s", err)
		} else {
			err = rtkcsm.ImportSave(file)
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

	var alertReader reader.AlertReader
	switch config.ReaderType {
	case "", "json":
		alertReader = &reader.JSONAlertReader{}
	case "suricata":
		alertReader = &reader.SuricataAlertReader{}
	case "zeek":
		alertReader = &reader.ZeekAlertReader{}
	case "suricata-tenzir":
		alertReader = &reader.SuricataTenzirAlertReader{}
	case "ocsf":
		alertReader = &reader.OCSFAlertReader{}
	default:
		panic("reader type is not known")
	}

	if config.VisualizationListenAddress != "" {
		go visualization.Start(config.VisualizationListenAddress, rtkcsm, assets)
	}

	var selectedTransport transport.Transport
	switch config.TransportType {
	case "", "file":
		if config.TransportFilePath != "" {
			selectedTransport = &transport.FileTransport{
				FilePath: config.TransportFilePath,
			}
		}
	case "tcp":
		selectedTransport = &transport.TcpTransport{
			ListenAddress: config.TransportListenAddress,
		}
	default:
		panic("transport type is not known")
	}

	if selectedTransport != nil {
		err := selectedTransport.Start(rtkcsm, alertReader)
		if err != nil {
			log.Panicf("error channeling alerts %s", err)
		}
	}

	rtkcsm.Stop(config.ExportFormats)

	endTime := time.Now()
	fmt.Printf("Start Time: %s \nEnd Time: %s\nExecution Time: %s\n", startTime, endTime, endTime.Sub(startTime))
	fmt.Printf("Generated graphs: %d\n", rtkcsm.GetGraphList(-1).Count)

	if profilerOptions.Has(structure.MemoryProfilerOptionFlag) || profilerOptions.Has(structure.AlertsProfilerOptionFlag) {
		file, err := os.Create("alerts-performance.csv")
		if err != nil {
			log.Panic(err)
		}

		structure.PerformanceManager.GetSeries("alerts").Save(file)
	}

	if profilerOptions.Has(structure.GraphRankingProfilerOptionFlag) {
		file, err := os.Create("graph-ranking.csv")
		if err != nil {
			log.Panic(err)
		}
		structure.PerformanceManager.GetSeries("graph-ranking").Save(file)
	}

	if profilerOptions.Has(structure.GraphNumberProfilerOptionFlag) {
		file, err := os.Create("graph-count.csv")
		if err != nil {
			log.Panic(err)
		}
		structure.PerformanceManager.GetSeries("graph-count").Save(file)
	}

	if config.VisualizationListenAddress != "" {
		<-make(chan struct{})
	}
}
