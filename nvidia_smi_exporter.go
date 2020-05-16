package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
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
		// TODO(glynternet): try getting units and add to description of each metric
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
			var unparseables []string
			var queryFieldUnsupporteds []string
			var unknownErrorMetrics []string
			var pstateUnparesable int
			name := fmt.Sprintf("%s[%s]", row[0], row[1])
			for idx, value := range row[2:] {
				v, knownErr, err := parseValue(value)
				metricName := metricList[idx]
				if metricName == "pstate" {
					v, err := parsePstate(value)
					if err != nil {
						_ = logger.Log(
							log.Message("Error parsing pstate level"),
							log.Error(err),
							log.KV{K: "index", V: idx},
							log.KV{K: "value", V: value},
							log.KV{K: "correspondingFlag", V: metricName},
						)
						pstateUnparesable++
						continue
					}
					writeMetric(logger, response, metricName, name, v)
					continue
				}
				if err != nil {
					_ = logger.Log(
						log.Message("Error parsing value for metric"),
						log.Error(err),
						log.KV{K: "index", V: idx},
						log.KV{K: "value", V: value},
						log.KV{K: "correspondingFlag", V: metricName},
					)
					unparseables = append(unparseables, metricName)
					continue
				}
				switch knownErr {
				case queryFieldUnsupported:
					queryFieldUnsupporteds = append(queryFieldUnsupporteds, metricName)
					continue
				case unknownError:
					unknownErrorMetrics = append(unknownErrorMetrics, metricName)
					continue
				}
				writeMetric(logger, response, metricName, name, v)
			}
			writeMetricCountWithLoggedValues(logger, response, "unparseable_query_result_value_count", name, unparseables)
			writeMetricCountWithLoggedValues(logger, response, "query_field_unsupported_count", name, queryFieldUnsupporteds)
			writeMetricCountWithLoggedValues(logger, response, "unknown_error_count", name, unknownErrorMetrics)
			writeMetric(logger, response, "pstate_unparseable", name, float64(pstateUnparesable))
		}
	}
}

func parsePstate(value string) (float64, error) {
	lenValue := len(value)
	if lenValue < 2 {
		return 0, fmt.Errorf("pstate value should be longer than 2 characters but is %d", lenValue)
	}
	firstChar := value[0]
	if firstChar != 'P' {
		return 0, fmt.Errorf("expected first character P but received: %c", firstChar)
	}
	level := value[1:]
	v, err := strconv.ParseInt(level, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("pstate level unparseable: %q", level)
	}
	return float64(v), nil
}

func writeMetric(logger log.Logger, w io.Writer, metricName, gpuName string, value float64) {
	if _, err := fmt.Fprintf(w, "nvidia_%s{gpu=\"%s\"} %f\n", metricName, gpuName, value); err != nil {
		_ = logger.Log(
			log.Message("Error writing response"),
			log.Error(err))
	}
}

func writeMetricCountWithLoggedValues(logger log.Logger, w io.Writer, metricName, gpuName string, values []string) {
	writeMetric(logger, w, metricName, gpuName, float64(len(values)))
	_ = logger.Log(
		log.Message("non-standard metric values"),
		log.KV{K: "metricName", V: metricName},
		log.KV{K: "gpuName", V: gpuName},
		log.KV{K: "values", V: values},
	)
}

type knownError string

const (
	queryFieldUnsupported = "metric is unsupported"
	unknownError          = "unknown error"
)

// returns parsed value, known error, or error
func parseValue(value string) (float64, knownError, error) {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v, "", nil
	}
	switch value {
	case "[Not Supported]":
		return 0, queryFieldUnsupported, nil
	case "Enabled":
		return 1, "", nil
	case "Disabled":
		return 0, "", nil
	case "[Unknown Error]":
		return 0, unknownError, nil
	}
	return 0, "", fmt.Errorf("unparsable query result value: %q", value)
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
