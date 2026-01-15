//go:build !production
package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manages online registry for Rockets !TODO",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Adding to registry... " + constants.ApplicationName)
		_, err := containers.Manager()
		if err != nil {

		}
	},
}

func init() {
	rootCmd.AddCommand(registryCmd)
}
