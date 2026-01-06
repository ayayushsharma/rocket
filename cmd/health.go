package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Launches specified application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		var conn containers.ContainerManager
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			os.Exit(1)
		}
		appName := viper.GetString("app")
		conn.StartService(appName)
		slog.Debug("Successfully started application", "application", appName)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.Flags().String("app", "", "application name to start")
}
