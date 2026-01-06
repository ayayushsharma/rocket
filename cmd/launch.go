package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launches specified application",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		var conn containers.ContainerManager
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			os.Exit(1)
		}
		appName := args[0]
		err = conn.StartService(appName)
		if err != nil {
			slog.Debug("Failed to start container. Exiting")
			os.Exit(1)
		}
		slog.Debug("Successfully started application", "application", appName)
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
