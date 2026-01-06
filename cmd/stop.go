package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [container_name]",
	Short: "Stops applications",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Stopping... " + constants.ApplicationName)
		var conn containers.ContainerManager
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			os.Exit(1)
		}
		appName := args[0]
		err = conn.StopService(appName)
		if err != nil {
			slog.Debug("Failed to start container. Exiting")
			os.Exit(1)
		}
		slog.Debug("Successfully stopped application", "application", appName)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
