package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "registrys specified containerised application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		_, err := containers.Manager()
		if err != nil {

		}
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
}
