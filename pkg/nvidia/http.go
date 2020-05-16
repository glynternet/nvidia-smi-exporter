package nvidia

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/glynternet/pkg/log"
)

const (
	metricNameUnparseableQueryResult = "unparseable_query_result_value"
	metricNameQueryFieldUnsupported  = "query_field_unsupported"
	metricNameUnknownError           = "unknown_error"
	metricNamePstateUnparseable      = "pstate_unparseable"
)

func MetricsHandler(logger log.Logger, executable string, fields []string) http.Handler {
	metricList := make([]string, len(fields))
	for i, field := range fields {
		metricList[i] = strings.Replace(field, ".", "_", -1)
	}
	args := []string{"--query-gpu=name,index," + strings.Join(fields, ","),
		// TODO(glynternet): try getting units and add to description of each metric
		"--format=csv,noheader,nounits"}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		out, err := exec.Command(executable, args...).Output()
		if err != nil {
			_ = logger.Log(
				log.Message("error executing nvidia-smi executable"),
				log.Error(err),
				log.KV{K: "executable", V: executable},
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
			problematicMetricValues := make(map[string][]string)
			gpuName := fmt.Sprintf("%s[%s]", row[0], row[1])
			for idx, value := range row[2:] {
				v, knownErr, err := parseValue(value)
				metricName := metricList[idx]
				if metricName == "pstate" {
					v, err := parsePstate(value)
					if err != nil {
						problematicMetricValues[metricNamePstateUnparseable] = append(problematicMetricValues[metricNamePstateUnparseable], metricName)
						continue
					}
					writeMetric(logger, response, metricName, gpuName, v)
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
					problematicMetricValues[metricNameUnparseableQueryResult] = append(problematicMetricValues[metricNameUnparseableQueryResult], metricName)
					continue
				}
				switch knownErr {
				case knownErrorQueryFieldUnsupported:
					problematicMetricValues[metricNameQueryFieldUnsupported] = append(problematicMetricValues[metricNameQueryFieldUnsupported], metricName)
					continue
				case knownErrorUnknownError:
					problematicMetricValues[metricNameUnknownError] = append(problematicMetricValues[metricNameUnknownError], metricName)
					continue
				}
				writeMetric(logger, response, metricName, gpuName, v)
			}
			for metricName, values := range problematicMetricValues {
				writeMetricCountWithLoggedValues(logger, response, metricName, gpuName, values)
			}
		}
	})
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
	if len(values) > 0 {
		// TODO(glynternet): make this only in debug logging
		_ = logger.Log(
			log.Message("non-standard metric values"),
			log.KV{K: "metricName", V: metricName},
			log.KV{K: "gpuName", V: gpuName},
			log.KV{K: "values", V: values},
		)
	}
}

type knownError int

const (
	// when a knownError is required but there is actually no error :)
	knownErrorNone knownError = iota
	knownErrorQueryFieldUnsupported
	// seems strange but it's a known case where nvidia-smi returns "Unknown Error"
	knownErrorUnknownError
)

// returns parsed value, known error, or error
func parseValue(value string) (float64, knownError, error) {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return v, knownErrorNone, nil
	}
	switch value {
	case "[Not Supported]":
		return 0, knownErrorQueryFieldUnsupported, nil
	case "Enabled":
		return 1, knownErrorNone, nil
	case "Disabled":
		return 0, knownErrorNone, nil
	case "[Unknown Error]":
		return 0, knownErrorUnknownError, nil
	}
	return 0, knownErrorNone, fmt.Errorf("unparsable query result value: %q", value)
}
