package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/workspace"
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister [container-name]",
	Short: "unregisters specified containerised application",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error){
		slog.Debug("UnRegistering... " + constants.ApplicationName)
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			return
		}
		containerName := args[0]
		err = unregisterApplication(conn, containerName)
		if err != nil {
			slog.Debug("Failed to unregister application", "error", err)
			return
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
	unregisterCmd.Flags().Bool("force", false, "to forcefully unregister an app")
}

func unregisterApplication(
	conn containers.ContainerManager,
	containerName string,
) (err error) {
	force := viper.GetBool("force")
	err = conn.StopService(containerName)
	if err != nil {
		slog.Debug("Failed to stop application", "error", err)
		if !force {
			return nil
		}
	}

	err = conn.RemoveContainer(containerName, force)
	if err != nil {
		slog.Debug("Failed to unregister application container", "error", err)
	}

	err = workspace.Unregister(containerName)
	if err != nil {
		slog.Debug("Failed to unregister app from workspace", "error", err)
		return
	}

	return nil
}
