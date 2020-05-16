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
	fields := defaultFields()
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
						log.KV{K: "value", V: value},
						log.KV{K: "correspondingFlag", V: metricList[idx]},
					)
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

func defaultFields() []string {
	return []string{
		"clocks.applications.gr",
		"clocks.applications.graphics",
		"clocks.applications.mem",
		"clocks.applications.memory",
		"clocks.current.graphics",
		"clocks.current.memory",
		"clocks.current.sm",
		"clocks.current.video",
		"clocks.default_applications.gr",
		"clocks.default_applications.graphics",
		"clocks.default_applications.mem",
		"clocks.default_applications.memory",
		"clocks.gr",
		"clocks.max.gr",
		"clocks.max.graphics",
		"clocks.max.mem",
		"clocks.max.memory",
		"clocks.max.sm",
		"clocks.mem",
		"clocks.sm",
		"clocks_throttle_reasons.gpu_idle",
		"clocks_throttle_reasons.hw_power_brake_slowdown",
		"clocks_throttle_reasons.hw_slowdown",
		"clocks_throttle_reasons.hw_thermal_slowdown",
		"clocks_throttle_reasons.sw_thermal_slowdown",
		"clocks_throttle_reasons.sync_boost",
		"clocks.video",
		"driver_version",
		"ecc.errors.corrected.aggregate.device_memory",
		"ecc.errors.corrected.aggregate.l1_cache",
		"ecc.errors.corrected.aggregate.l2_cache",
		"ecc.errors.corrected.aggregate.register_file",
		"ecc.errors.corrected.aggregate.texture_memory",
		"ecc.errors.corrected.aggregate.total",
		"ecc.errors.corrected.volatile.l1_cache",
		"ecc.errors.corrected.volatile.l2_cache",
		"ecc.errors.corrected.volatile.register_file",
		"ecc.errors.corrected.volatile.texture_memory",
		"ecc.errors.corrected.volatile.total",
		"ecc.errors.uncorrected.aggregate.device_memory",
		"ecc.errors.uncorrected.aggregate.l1_cache",
		"ecc.errors.uncorrected.aggregate.l2_cache",
		"ecc.errors.uncorrected.aggregate.register_file",
		"ecc.errors.uncorrected.aggregate.texture_memory",
		"ecc.errors.uncorrected.aggregate.total",
		"ecc.errors.uncorrected.volatile.device_memory",
		"ecc.errors.uncorrected.volatile.l1_cache",
		"ecc.errors.uncorrected.volatile.l2_cache",
		"ecc.errors.uncorrected.volatile.register_file",
		"ecc.errors.uncorrected.volatile.texture_memory",
		"ecc.errors.uncorrected.volatile.total",
		"encoder.stats.averageFps",
		"encoder.stats.averageLatency",
		"encoder.stats.sessionCount",
		"enforced.power.limit",
		"fan.speed",
		"memory.free",
		"memory.total",
		"memory.used",
		"power.default_limit",
		"power.draw",
		"power.limit",
		"power.management",
		"power.max_limit",
		"power.min_limit",
		"pstate",
		"retired_pages.dbe",
		"retired_pages.double_bit.count",
		"retired_pages.pending",
		"retired_pages.sbe",
		"retired_pages.single_bit_ecc.count",
		"temperature.gpu",
		"utilization.gpu",
		"utilization.memory"}
}

func envOrDefault(key string, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}
