package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/glynternet/pkg/log"
)

// name, index, temperature.gpu, utilization.gpu,
// utilization.memory, memory.total, memory.free, memory.used

func main() {
	var (
		listenAddress string
		metricsPath   string
	)

	flag.StringVar(&listenAddress, "web.listen-address", ":9101", "Address to listen on")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.Parse()

	logger := log.NewLogger()
	http.HandleFunc(metricsPath, metrics(logger))
	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		_ = logger.Log(
			log.Message("Error running ListenAndServe"),
			log.Error(err))
	}
}

func metrics(logger log.Logger) func(http.ResponseWriter, *http.Request) {
	command := envOrDefault("NVIDIA_SMI", "nvidia-smi")
	fields := []string{
		"driver_version",
		"temperature.gpu",
		"utilization.gpu",
		"utilization.memory",
		"memory.total",
		"memory.free",
		"memory.used",
		"fan.speed",
		"power.draw",
		"clocks.current.graphics",
		"clocks.current.sm",
		"clocks.current.memory",
		"clocks.current.video",
		"encoder.stats.sessionCount",
		"encoder.stats.averageFps",
		"encoder.stats.averageLatency"}
	metricList := make([]string, len(fields))
	for i, field := range fields {
		metricList[i] = strings.Replace(field, ".", "_", -1)
	}
	args := []string{"--query-gpu=name,index," + strings.Join(fields, ","),
		// TODO(glynternet): try getting units and adding to description of each metric
		"--format=csv,noheader,nounits"}
	return func(response http.ResponseWriter, request *http.Request) {
		out, err := exec.Command(command, args...).Output()
		if err != nil {
			_ = logger.Log(
				log.Message("error executing nvidia-smi command"),
				log.Error(err),
				log.KV{K: "command", V: command},
				log.KV{K: "args", V: args},
			)
		}

		csvReader := csv.NewReader(bytes.NewReader(out))
		csvReader.TrimLeadingSpace = true
		records, err := csvReader.ReadAll()
		if err != nil {
			_ = logger.Log(
				log.Message("error reading CSV"),
				log.Error(err))
			return
		}

		for _, row := range records {
			var unsupported int
			name := fmt.Sprintf("%s[%s]", row[0], row[1])
			for idx, value := range row[2:] {
				v, err := strconv.ParseFloat(value, 64)
				if err != nil {
					_ = logger.Log(
						log.Message("Error parsing value for metric"),
						log.Error(err),
						log.KV{K: "index", V: idx},
						log.KV{K: "value", V: value})
					unsupported++
					continue
				}
				if _, err := fmt.Fprintf(response, "nvidia_%s{gpu=\"%s\"} %f\n", metricList[idx], name, v); err != nil {
					_ = logger.Log(
						log.Message("Error writing response"),
						log.Error(err))
				}
			}
			if _, err := fmt.Fprintf(response, "nvidia_unsupported_metrics_count{gpu=\"%s\"} %d\n", name, unsupported); err != nil {
				_ = logger.Log(
					log.Message("Error writing response"),
					log.Error(err))
			}
		}
	}
}

func envOrDefault(key string, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}
