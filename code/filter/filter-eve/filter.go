package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/alexflint/go-arg"
)

type Dataset struct {
	filePath string
	label    string
}

type DatasetEntry struct {
	data  string
	label string
}

type Worker[T any] struct {
	mutex      *sync.RWMutex
	isActive   bool
	workerFunc func(line T)
}

func NewWorker[T any](workerFunc func(line T)) *Worker[T] {
	return &Worker[T]{
		mutex:      &sync.RWMutex{},
		isActive:   false,
		workerFunc: workerFunc,
	}
}

func (w *Worker[T]) SetActive(isActive bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.isActive = isActive
}

func (w *Worker[T]) IsActive() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.isActive
}

func (w *Worker[T]) Start(input chan T) Worker[T] {
	for {
		w.SetActive(false)
		line := <-input
		w.SetActive(true)
		w.workerFunc(line)
	}
}

var exclusions = []*regexp.Regexp{
	regexp.MustCompile("^SURICATA.*$"),
	regexp.MustCompile("^ET INFO.*$"),
	regexp.MustCompile("^ET DNS Query.*$"),
	regexp.MustCompile("^ET JA3.*$"),
	regexp.MustCompile("^ET USER_AGENTS.*$"),
}

var workerThreads = 10000
var datasets = []Dataset{
	{filePath: "alerts/benign.pcap.json", label: "benign"},
	{filePath: "alerts/botnet.pcap.json", label: "botnet"},
	{filePath: "alerts/botnet.pcap.json", label: "botnet"},
	{filePath: "alerts/ddos-1-loic-http.pcap.json", label: "ddos-1-loic-http"},
	{filePath: "alerts/ddos-2-loic-udp-1.pcap.json", label: "ddos-2-loic-udp-1"},
	{filePath: "alerts/ddos-2-loic-udp-2.pcap.json", label: "ddos-2-loic-udp-2"},
	{filePath: "alerts/ddos-3-hoic.pcap.json", label: "ddos-3-hoic"},
	{filePath: "alerts/dos-1-hulk.pcap.json", label: "dos-1-hulk"},
	{filePath: "alerts/dos-2-goldeneye.pcap.json", label: "dos-2-goldeneye"},
	{filePath: "alerts/dos-3-slowloris.pcap.json", label: "dos-3-slowloris"},
	{filePath: "alerts/ftp-bruteforce-1.pcap.json", label: "ftp-bruteforce-1"},
	{filePath: "alerts/ftp-bruteforce-2.pcap.json", label: "ftp-bruteforce-2"},
	{filePath: "alerts/ftp-bruteforce-3.pcap.json", label: "ftp-bruteforce-3"},
	{filePath: "alerts/multi-stage-dropbox-download-1.pcap.json", label: "multi-stage-dropbox-download-1"},
	{filePath: "alerts/multi-stage-dropbox-download-2.pcap.json", label: "multi-stage-dropbox-download-2"},
	{filePath: "alerts/multi-stage-infiltration-1.pcap.json", label: "multi-stage-infiltration-1"},
	{filePath: "alerts/multi-stage-infiltration-2.pcap.json", label: "multi-stage-infiltration-2"},
	{filePath: "alerts/multi-stage-nmap-1.pcap.json", label: "multi-stage-nmap-1"},
	{filePath: "alerts/multi-stage-nmap-2.pcap.json", label: "multi-stage-nmap-2"},
	{filePath: "alerts/sql-injection-1.pcap.json", label: "sql-injection-1"},
	{filePath: "alerts/sql-injection-2.pcap.json", label: "sql-injection-2"},
	{filePath: "alerts/ssh-bruteforce-1-patator.pcap.json", label: "ssh-bruteforce-1-patator"},
	{filePath: "alerts/web-bruteforce-1.pcap.json", label: "web-bruteforce-1"},
	{filePath: "alerts/web-bruteforce-2.pcap.json", label: "web-bruteforce-2"},
	{filePath: "alerts/xss-1.pcap.json", label: "xss-1"},
	{filePath: "alerts/xss-2.pcap.json", label: "xss-2"},
}

type Arguments struct {
	OutputFile string `arg:"--output"`
}

func main() {
	var arguments Arguments
	arg.MustParse(&arguments)

	outputFile, err := os.OpenFile(arguments.OutputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer outputFile.Close()
	outputFileMutex := &sync.Mutex{}

	input := make(chan DatasetEntry, 8192)
	filter := make(chan map[string]interface{}, 8192)
	writer := make(chan map[string]interface{}, 8192)

	processDataset := func(dataset Dataset) {
		file, err := os.Open(dataset.filePath)
		if err != nil {
			log.Panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			data := scanner.Text()
			input <- DatasetEntry{
				data:  data,
				label: dataset.label,
			}
		}
	}

	parserWorkers := []*Worker[DatasetEntry]{}

	counterEvents := 0
	counterEventsBytes := 0
	counterAlerts := 0
	counterAlertsBytes := 0
	counterSync := &sync.RWMutex{}

	printProgress := func() {
		counterSync.RLock()
		fmt.Printf("\rProcessed events: %d (%.4f GB) -> %d (%.4f GB)", counterEvents, float64(counterEventsBytes)/1024/1024/1024, counterAlerts, float64(counterAlertsBytes)/1024/1024/1024)
		counterSync.RUnlock()
	}

	go func() {
		for {
			printProgress()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	// Parser workers
	for i := 0; i < workerThreads; i++ {
		worker := NewWorker(func(entry DatasetEntry) {
			counterSync.Lock()
			counterEvents += 1
			counterEventsBytes += len([]byte(entry.data))
			counterSync.Unlock()

			var obj map[string]interface{}

			err := json.Unmarshal([]byte(entry.data), &obj)
			if err != nil {
				log.Panicf("%s: %v\n%v", entry.label, err, string(entry.data))
			} else {
				obj["label"] = entry.label
				filter <- obj
			}
		})
		parserWorkers = append(parserWorkers, worker)
		go worker.Start(input)
	}

	filterWorkers := []*Worker[map[string]interface{}]{}

	// Filter workers
	for i := 0; i < workerThreads; i++ {
		worker := NewWorker(func(obj map[string]interface{}) {
			if alert, ok := obj["alert"]; ok {
				alert := alert.(map[string]interface{})
				signature, ok := alert["signature"].(string)
				if ok {
					for _, exclusion := range exclusions {
						if exclusion.MatchString(signature) {
							return
						}
					}

					writer <- obj
				}
			}
		})
		filterWorkers = append(filterWorkers, worker)
		go worker.Start(filter)
	}

	// Writer workers
	for i := 0; i < workerThreads; i++ {
		worker := NewWorker(func(obj map[string]interface{}) {
			// Write to file
			newLine, err := json.Marshal(&obj)
			if err != nil {
				log.Panic(err)
			}

			newLineStr := string(newLine) + "\n"
			counterSync.Lock()
			counterAlerts += 1
			counterAlertsBytes += len([]byte(newLineStr))
			counterSync.Unlock()

			outputFileMutex.Lock()
			_, err = outputFile.WriteString(newLineStr)
			if err != nil {
				log.Panic(err)
			}
			outputFileMutex.Unlock()
		})
		filterWorkers = append(filterWorkers, worker)
		go worker.Start(writer)
	}

	wg := &sync.WaitGroup{}
	for _, dataset := range datasets {
		wg.Add(1)
		go func() {
			processDataset(dataset)
			wg.Done()
			log.Printf("\rDone with: %s\n", dataset.label)
		}()
	}

	wg.Wait()

	for {
		printProgress()

		if len(input) == 0 {
			isOneWorkerActive := false
			for _, worker := range parserWorkers {
				if worker.IsActive() {
					isOneWorkerActive = true
					break
				}
			}

			for _, worker := range filterWorkers {
				if worker.IsActive() {
					isOneWorkerActive = true
					break
				}
			}

			if !isOneWorkerActive {
				break
			}
		}

		time.Sleep(time.Millisecond * 100)
	}
}
