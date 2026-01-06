package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/register"
	"log/slog"
	// "os"

	"github.com/spf13/cobra"
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "unregisters specified containerised application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("UnRegistering... " + constants.ApplicationName)
		// conn, err := containers.ConnectPodman()
		// if err != nil {
		// 	slog.Debug("Failed to select application", "error", err)
		// 	os.Exit(1)
		// }
		//
		// appToRegister, err := registry.ApplicationToRegister()
		// if err != nil {
		// 	slog.Debug("Failed to select application", "error", err)
		// 	os.Exit(1)
		// }
		//
	},
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
}

func unregisterApplication(
	conn containers.Container,
	containerName string,
) (err error) {
	err = conn.StopService(containerName)
	if err != nil {
		slog.Debug("Failed to start application", "error", err)
		return
	}

	err = conn.RemoveContainer("", true)
	if err != nil {
		slog.Debug("Failed to unregister application container", "error", err)
		return
	}

	err = register.UnregisterApplicationToConf(containerName)
	if err != nil {
		slog.Debug(
			"Failed to unregister application container to configuration",
			"error",
			err,
		)
		return
	}

	err = register.RefreshRouterConf()
	if err != nil {
		slog.Debug("Failed to unregister application to routes", "error", err)
		return
	}
	return nil
}
