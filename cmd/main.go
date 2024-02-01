package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"

	"atomicgo.dev/cursor"
)

func main() {
	pipelineInfo := getPipelineInfo()
	fmt.Println("Logstash Top")
	fmt.Println("--------------------------")
	area := cursor.NewArea()
	header := "Pipeline\n"
	area.Update(header)

	origContent := header
	for {
		content := requestContent(pipelineInfo)
		area.Update(origContent + content)
		time.Sleep(1 * time.Second)
	}

}

type FlowIntervals struct {
	Current       float64 `json:"current"`
	Last1Minute   float64 `json:"last_1_minute"`
	Last5Minutes  float64 `json:"last_5_minutes"`
	Last15Minutes float64 `json:"last_15_minutes"`
	Last1Hour     float64 `json:"last_1_hour"`
	Lifetime      float64 `json:"lifetime"`
}

type FlowWorkers struct {
	WorkerUtilization    FlowIntervals `json:"worker_utilization"`
	WorkerMillisPerEvent FlowIntervals `json:"worker_millis_per_event"`
}

type FilterOrOutput struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Events struct {
		In               int `json:"in"`
		Out              int `json:"out"`
		DurationInMillis int `json:"duration_in_millis"`
	} `json:"events"`
	Flow FlowWorkers `json:"flow"`
}

type PipelineData struct {
	Events struct {
		QueuePushDurationInMillis int `json:"queue_push_duration_in_millis"`
		In                        int `json:"in"`
		Filtered                  int `json:"filtered"`
		Out                       int `json:"out"`
		DurationInMillis          int `json:"duration_in_millis"`
	} `json:"events"`
	Flow struct {
		QueuePersistedGrowthEvents FlowIntervals `json:"queue_persisted_growth_events"`
		QueuePersistedGrowthBytes  FlowIntervals `json:"queue_persisted_growth_bytes"`
		OutputThroughput           FlowIntervals `json:"output_throughput"`
		QueueBackpressure          FlowIntervals `json:"queue_backpressure"`
		InputThroughput            FlowIntervals `json:"input_throughput"`
		FilterThroughput           FlowIntervals `json:"filter_throughput"`
		WorkerConcurrency          FlowIntervals `json:"worker_concurrency"`
	} `json:"flow"`
	Plugins struct {
		Inputs []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Events struct {
				QueuePushDurationInMillis int `json:"queue_push_duration_in_millis"`
				Out                       int `json:"out"`
			} `json:"events"`
			Flow struct {
				Throughput FlowIntervals `json:"throughput"`
			} `json:"flow"`
		} `json:"inputs"`
		Codecs []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Encode struct {
				WritesIn         int `json:"writes_in"`
				DurationInMillis int `json:"duration_in_millis"`
			} `json:"encode"`
			Decode struct {
				WritesIn         int `json:"writes_in"`
				Out              int `json:"out"`
				DurationInMillis int `json:"duration_in_millis"`
			} `json:"decode"`
		} `json:"codecs"`
		Filters []FilterOrOutput `json:"filters"`
		Outputs []FilterOrOutput `json:"outputs"`
	} `json:"plugins"`
	Reloads struct {
		LastFailureTimestamp any `json:"last_failure_timestamp"`
		Failures             int `json:"failures"`
		LastSuccessTimestamp any `json:"last_success_timestamp"`
		Successes            int `json:"successes"`
		LastError            any `json:"last_error"`
	} `json:"reloads"`
	Queue struct {
		Data struct {
			Path             string `json:"path"`
			FreeSpaceInBytes int64  `json:"free_space_in_bytes"`
			StorageType      string `json:"storage_type"`
		} `json:"data"`
		Events   int `json:"events"`
		Capacity struct {
			MaxQueueSizeInBytes int `json:"max_queue_size_in_bytes"`
			QueueSizeInBytes    int `json:"queue_size_in_bytes"`
			MaxUnreadEvents     int `json:"max_unread_events"`
			PageCapacityInBytes int `json:"page_capacity_in_bytes"`
		} `json:"capacity"`
		Type                string `json:"type"`
		EventsCount         int    `json:"events_count"`
		QueueSizeInBytes    int    `json:"queue_size_in_bytes"`
		MaxQueueSizeInBytes int    `json:"max_queue_size_in_bytes"`
	} `json:"queue"`
	Hash        string `json:"hash"`
	EphemeralID string `json:"ephemeral_id"`
}

// Generated with https://mholt.github.io/json-to-go/
type PipelineAnswer struct {
	Host        string `json:"host"`
	Version     string `json:"version"`
	HTTPAddress string `json:"http_address"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	EphemeralID string `json:"ephemeral_id"`
	Status      string `json:"status"`
	Snapshot    bool   `json:"snapshot"`
	Pipeline    struct {
		Workers    int `json:"workers"`
		BatchSize  int `json:"batch_size"`
		BatchDelay int `json:"batch_delay"`
	} `json:"pipeline"`
	Pipelines map[string]PipelineData `json:"pipelines"`
}

type PipelineConfiguration struct {
	EphemeralID            string `json:"ephemeral_id"`
	Hash                   string `json:"hash"`
	Workers                int    `json:"workers"`
	BatchSize              int    `json:"batch_size"`
	BatchDelay             int    `json:"batch_delay"`
	ConfigReloadAutomatic  bool   `json:"config_reload_automatic"`
	ConfigReloadInterval   int64  `json:"config_reload_interval"`
	DeadLetterQueueEnabled bool   `json:"dead_letter_queue_enabled"`
}

type NodeOverview struct {
	Host        string `json:"host"`
	Version     string `json:"version"`
	HTTPAddress string `json:"http_address"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	EphemeralID string `json:"ephemeral_id"`
	Status      string `json:"status"`
	Snapshot    bool   `json:"snapshot"`
	Pipeline    struct {
		Workers    int `json:"workers"`
		BatchSize  int `json:"batch_size"`
		BatchDelay int `json:"batch_delay"`
	} `json:"pipeline"`
	Pipelines map[string]PipelineConfiguration `json:"pipelines"`
	Os        struct {
		Name                string `json:"name"`
		Arch                string `json:"arch"`
		Version             string `json:"version"`
		AvailableProcessors int    `json:"available_processors"`
	} `json:"os"`
	Jvm struct {
		Pid               int    `json:"pid"`
		Version           string `json:"version"`
		VMVersion         string `json:"vm_version"`
		VMVendor          string `json:"vm_vendor"`
		VMName            string `json:"vm_name"`
		StartTimeInMillis int64  `json:"start_time_in_millis"`
		Mem               struct {
			HeapInitInBytes    int `json:"heap_init_in_bytes"`
			HeapMaxInBytes     int `json:"heap_max_in_bytes"`
			NonHeapInitInBytes int `json:"non_heap_init_in_bytes"`
			NonHeapMaxInBytes  int `json:"non_heap_max_in_bytes"`
		} `json:"mem"`
		GcCollectors []string `json:"gc_collectors"`
	} `json:"jvm"`
}

var yellow = color.New(color.FgYellow)
var red = color.New(color.FgRed)
var green = color.New(color.FgGreen)
var blue = color.New(color.FgBlue)

func getDiffString(diff float64) string {
	diffOutput := ""
	v := fmt.Sprintf("%8.4f", diff)
	if diff < -0.1 {
		diffOutput += red.Sprintf("%15s", v)
	} else if diff > 0.1 {
		diffOutput += green.Sprintf("%15s", v)
	} else {
		diffOutput += fmt.Sprintf("%15s", v)
	}
	return diffOutput
}

func printFlowWorkers(fw []FilterOrOutput) string {
	output := fmt.Sprintf("+----------+---------------+---------------+---------------+---------------+\n")
	output += fmt.Sprintf("|%-10s|%-15s|%-15s|%-15s|%-15s|\n", "Name", "CMillisPerEvent", "CurrentWU", "DiffLifetime", "Diff1Minutes")
	output += fmt.Sprintf("+----------+---------------+---------------+---------------+---------------+\n")
	for i := 0; i < len(fw); i++ {

		CurrentWU := fmt.Sprintf("%f", fw[i].Flow.WorkerUtilization.Current)
		CurrentWUMillis := fmt.Sprintf("%f", fw[i].Flow.WorkerMillisPerEvent.Current)

		output += fmt.Sprintf("|%-10s|%-15s|%-15s|%-15s|%-15s|\n",
			fw[i].Name,
			CurrentWUMillis,
			CurrentWU,
			getDiffString(fw[i].Flow.WorkerUtilization.Current-
				fw[i].Flow.WorkerUtilization.Lifetime),
			getDiffString(fw[i].Flow.WorkerUtilization.Current-
				fw[i].Flow.WorkerUtilization.Last1Minute),
		)
	}
	output += fmt.Sprintf("+----------+---------------+---------------+---------------+---------------+\n")
	return output

}

func getPipelineInfo() NodeOverview {
	resp, err := http.Get("http://localhost:9600/_node")
	if err != nil {
		log.Panicf("Error %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Error %s", err)
		os.Exit(1)
	}
	nodeOverview := NodeOverview{}
	err = json.Unmarshal(body, &nodeOverview)

	return nodeOverview

}

func requestContent(pipelineInfo NodeOverview) string {
	resp, err := http.Get("http://localhost:9600/_node/stats/pipelines")
	if err != nil {
		log.Panicf("Error %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("Error %s", err)
		os.Exit(1)
	}
	output := ""

	answer := PipelineAnswer{}
	err = json.Unmarshal(body, &answer)
	if err != nil {
		log.Panicf("Could not unmarshal json %s", err)
		os.Exit(1)
	}
	pipelineName := "test-http"

	limit := min(len(body)-1, 10000)
	output += fmt.Sprintf("%d %d\n", len(body), limit)
	output += fmt.Sprintf("Workers: %d | Batch Size: %d | In: %d | Filtered: %d | Out %d \n",
		pipelineInfo.Pipelines[pipelineName].Workers,
		pipelineInfo.Pipelines[pipelineName].BatchSize,
		answer.Pipelines[pipelineName].Events.In,
		answer.Pipelines[pipelineName].Events.Filtered,
		answer.Pipelines[pipelineName].Events.Out,
	)

	output += fmt.Sprintf("%30s", "===Inputs===\n")
	output += fmt.Sprintf("+----------+---------------+---------------+---------------+\n")
	output += fmt.Sprintf("|%-10s|%-15s|%-15s|%-15s|\n", "Name", "Current", "DiffLifetime", "Diff1Minutes")
	output += fmt.Sprintf("+----------+---------------+---------------+---------------+\n")
	for i := 0; i < len(answer.Pipelines[pipelineName].Plugins.Inputs); i++ {

		Current := fmt.Sprintf("%f", answer.Pipelines[pipelineName].Plugins.Inputs[i].Flow.Throughput.Current)
		output += fmt.Sprintf("|%-10s|%-15s|%-15s|%-15s|\n",
			answer.Pipelines[pipelineName].Plugins.Inputs[i].Name,
			Current,
			getDiffString(answer.Pipelines[pipelineName].Plugins.Inputs[i].Flow.Throughput.Current-answer.Pipelines[pipelineName].Plugins.Inputs[i].Flow.Throughput.Lifetime),
			getDiffString(answer.Pipelines[pipelineName].Plugins.Inputs[i].Flow.Throughput.Current-answer.Pipelines[pipelineName].Plugins.Inputs[i].Flow.Throughput.Last1Minute),
		)
	}
	output += fmt.Sprintf("+----------+---------------+---------------+---------------+\n")

	output += fmt.Sprintf("%30s", "===Filters===\n")
	output += printFlowWorkers(answer.Pipelines[pipelineName].Plugins.Filters)

	output += fmt.Sprintf("%30s", "===Outputs===\n")
	output += printFlowWorkers(answer.Pipelines[pipelineName].Plugins.Outputs)

	// output += fmt.Sprintf("%f", answer.Pipelines[pipelineName].Flow.QueuePersistedGrowthEvents.Last1Minute)

	// for i := 0; i < limit; i += 100 {
	// 	output += string(body[i:min(i+100, limit-1)]) + "\n"
	// }
	return output
}
