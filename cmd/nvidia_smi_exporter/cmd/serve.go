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
					getQueryFields())), "error running ListenAndServe")
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
