package main

import (
	"io"
	"os"

	"github.com/glynternet/nvidia-smi-exporter/cmd/nvidia-smi-exporter/cmd"
	pkgcmd "github.com/glynternet/pkg/cmd"
	"github.com/glynternet/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func buildCmdTree(logger log.Logger, out io.Writer, rootCmd *cobra.Command) {
	rootCmd.AddCommand(pkgcmd.NewBashCompletion(rootCmd, out))
	for _, addCmd := range []func(log.Logger, io.Writer, *cobra.Command) error{
		cmd.Serve,
		cmd.ListMetrics,
	} {
		err := addCmd(logger, out, rootCmd)
		if err != nil {
			_ = logger.Log(
				log.Message("Error adding subcommand"),
				log.Error(err))
			os.Exit(1)
		}
	}

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		_ = logger.Log(
			log.Message("Error binding PFlags"),
			log.Error(err))
		os.Exit(1)
	}
}
