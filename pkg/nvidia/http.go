package nvidia

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/glynternet/pkg/log"
)

var (
	// extra metrics for context around parsing of nvidia-smi output
	metricNameUnparseableQueryResult = metricName("unparseable_query_result_value")
	metricNameQueryFieldUnsupported  = metricName("query_field_unsupported")
	metricNameUnknownError           = metricName("unknown_error")
	metricNamePstateUnparseable      = metricName("pstate_unparseable")

	// special parsing case
	metricNamePstate = metricName("pstate")
)

// MetricNames provides all of the metrics that can be produced by the exporter.
func MetricNames(smiFields []string) []string {
	mns := append(
		smiMetricNames(smiFields),
		metricNamePstateUnparseable,
		metricNameQueryFieldUnsupported,
		metricNameUnknownError,
		metricNameUnparseableQueryResult,
	)
	sort.Strings(mns)
	return mns
}

// smiMetricNames provides the names of the metrics specifically for the SMI query fields used
func smiMetricNames(fields []string) []string {
	ns := make([]string, len(fields))
	for i, field := range fields {
		ns[i] = metricName(strings.Replace(field, ".", "_", -1))
	}
	return ns
}

func metricName(suffix string) string {
	return `nvidia_` + suffix
}

func MetricsHandler(logger log.Logger, executable string, smiQueryFields []string) http.Handler {
	args := []string{"--query-gpu=name,index," + strings.Join(smiQueryFields, ","),
		// TODO(glynternet): try getting units and add to description of each metric
		"--format=csv,noheader,nounits"}
	metrics := smiMetricNames(smiQueryFields)
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		out, err := exec.Command(executable, args...).Output()
		if err != nil {
			_ = logger.Log(
				log.Message("error executing nvidia-smi executable"),
				log.ErrorMessage(err),
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
				log.ErrorMessage(err))
			return
		}

		for _, row := range records {
			problematicMetricValues := problematicMetricStore()
			gpuName := fmt.Sprintf("%s[%s]", row[0], row[1])
			for idx, value := range row[2:] {
				v, knownErr, err := parseValue(value)
				metricName := metrics[idx]
				if metricName == metricNamePstate {
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
						log.ErrorMessage(err),
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

// zero length slice for each metric so that we produce a metric even when there are no errors
func problematicMetricStore() map[string][]string {
	return map[string][]string{
		metricNameUnparseableQueryResult: nil,
		metricNameQueryFieldUnsupported:  nil,
		metricNameUnknownError:           nil,
		metricNamePstateUnparseable:      nil,
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
	if _, err := fmt.Fprintf(w, "%s{gpu=\"%s\"} %f\n", metricName, gpuName, value); err != nil {
		_ = logger.Log(
			log.Message("Error writing response"),
			log.ErrorMessage(err))
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
