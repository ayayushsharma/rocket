package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"

	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launches specified application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		containers.Connect()
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
