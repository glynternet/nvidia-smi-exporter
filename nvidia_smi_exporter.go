package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

// name, index, temperature.gpu, utilization.gpu,
// utilization.memory, memory.total, memory.free, memory.used

var (
	listenAddress string
	metricsPath   string
)

func metrics(response http.ResponseWriter, request *http.Request) {
	out, err := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,index,temperature.gpu,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,fan.speed,power.draw,clocks.current.graphics,clocks.current.sm,clocks.current.memory,clocks.current.video,encoder.stats.sessionCount,encoder.stats.averageFps,encoder.stats.averageLatency",
		"--format=csv,noheader,nounits").Output()

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	csvReader := csv.NewReader(bytes.NewReader(out))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	metricList := []string{
		"temperature_gpu", "utilization_gpu",
		"utilization_memory", "memory_total", "memory_free", "memory_used", "fan_speed", "power_draw",
		"clocks_current_graphics", "clocks_current_sm", "clocks_current_memory", "clocks_current_video",
		"encoder_stats_session_count", "encoder_stats_average_fps", "encoder_stats_average_latency",
	}

	for _, row := range records {
		name := fmt.Sprintf("%s[%s]", row[0], row[1])
		for idx, value := range row[2:] {
			fmt.Fprintf(response, "nvidia_%s{gpu=\"%s\"} %s\n", metricList[idx], name, value)
		}
	}
}

func init() {
	flag.StringVar(&listenAddress, "web.listen-address", ":9101", "Address to listen on")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.Parse()
}

func main() {
	//    addr := ":9101"
	//    if len(os.Args) > 1 {
	//        addr = ":" + os.Args[1]
	//    }

	http.HandleFunc(metricsPath, metrics)
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
