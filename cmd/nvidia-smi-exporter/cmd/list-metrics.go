package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/glynternet/nvidia-smi-exporter/pkg/nvidia"
	"github.com/glynternet/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ListMetrics(logger log.Logger, w io.Writer, parent *cobra.Command) error {
	var (
		cmd = &cobra.Command{
			Use:   "list-metric-names",
			Short: "list the metric names that would be produced",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, oscFiles []string) error {
				fmt.Println(strings.Join(nvidia.MetricNames(getQueryFields()), "\t"))
				return nil
			},
		}
	)

	parent.AddCommand(cmd)
	return errors.Wrap(viper.BindPFlags(cmd.Flags()), "binding pflags")
}
