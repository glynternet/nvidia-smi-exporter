package cmd

import (
	"io"
	"net/http"
	"os"

	"github.com/glynternet/nvidia_smi_exporter/pkg/nvidia"
	"github.com/glynternet/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Serve(logger log.Logger, w io.Writer, parent *cobra.Command) error {
	var (
		listenAddress string
		metricsPath   string
		cmd           = &cobra.Command{
			Use:  "serve GPU metrics",
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, oscFiles []string) error {
				return errors.Wrap(http.ListenAndServe(listenAddress, nvidia.MetricsHandler(logger,
					envOrDefault("NVIDIA_SMI", "nvidia-smi"),
					defaultFields())), "error running ListenAndServe")
			},
		}
	)

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&listenAddress, "web.listen-address", ":9101", "Address to listen on")
	cmd.Flags().StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	return errors.Wrap(viper.BindPFlags(cmd.Flags()), "binding pflags")
}

func envOrDefault(key string, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
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
