package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

type Config struct {
	LogFile        string `yaml:"log_file"`
	LogPattern     string `yaml:"log_pattern"`
	ExporterPort   string `yaml:"exporter_port"`
	RepeatInterval int    `yaml:"repeat_interval"`
	StateFile      string `yaml:"state_file"`
}

type State struct {
	LastLineOffset int64 `yaml:"last_line_offset"`
}

var (
	oomEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "oom_events_total",
			Help: "Total number of OOM events",
		},
		[]string{"pid", "process_name"},
	)
)

func main() {
	prometheus.MustRegister(oomEvents)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	done := make(chan struct{})
	go startLogMonitor(config, done)

	port := config.ExporterPort
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting OOM Exporter on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadOrCreateState(filename string) *State {
	state := &State{}
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return state
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading state file: %v", err)
	}

	if err := yaml.Unmarshal(data, state); err != nil {
		log.Fatalf("Error parsing state file: %v", err)
	}

	return state
}

func saveState(filename string, state *State) {
	data, err := yaml.Marshal(state)
	if err != nil {
		log.Fatalf("Error marshaling state: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatalf("Error writing state file: %v", err)
	}
}

func startLogMonitor(config *Config, done chan struct{}) {
	state := loadOrCreateState(config.StateFile)

	for {
		err := monitorLog(config, state)
		if err != nil {
			log.Printf("Error monitoring log: %v", err)
		}

		saveState(config.StateFile, state)

		select {
		case <-time.After(time.Duration(config.RepeatInterval) * time.Second):
		case <-done:
			return
		}
	}
}

func monitorLog(config *Config, state *State) error {
	file, err := os.Open(config.LogFile)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Если позиция в файле больше размера файла, сбросить позицию на начало
	fileInfo, _ := file.Stat()
	if state.LastLineOffset > fileInfo.Size() {
		state.LastLineOffset = 0
	}

	file.Seek(state.LastLineOffset, io.SeekStart)
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(config.LogPattern)

	now := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		// Проверка, что строка достаточно длинна
		if len(line) < 15 {
			continue
		}
		// Получение временной метки из строки лога
		timestampStr := line[:15]
		timestamp, err := time.Parse("Jan 02 15:04:05", timestampStr)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			continue
		}

		// Анализ события OOM
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 3 {
				pid := matches[1]
				processName := matches[2]
				oomEvents.WithLabelValues(pid, processName).Inc()

				fmt.Printf("[%s] Detected OOM event at %s: Process %s (PID: %s)\n", now.Format("2006-01-02 15:04:05"), timestamp, processName, pid)
			}
		}

		state.LastLineOffset, _ = file.Seek(0, io.SeekCurrent)
	}

	if state.LastLineOffset != 0 {
		fmt.Printf("[%s] No OOM events detected since last check.\n", now.Format("2006-01-02 15:04:05"))
	}

	return scanner.Err()
}
