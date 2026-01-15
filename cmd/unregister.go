package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/workspace"
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister [container-name]",
	Short: "Removes specified containerised application",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		slog.Debug("Unregistering... " + constants.ApplicationName)
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			return
		}
		containerName := args[0]
		err = unregisterApplication(conn, common.CompleteAppName(containerName))
		if err != nil {
			slog.Debug("Failed to unregister application", "error", err)
			return
		}

		return nil
	},
	ValidArgsFunction: unregisterAppCompletionFn,
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
	unregisterCmd.Flags().Bool(
		"force",
		false,
		"to forcefully unregister an app. Not recommended if other apps can use the image",
	)
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

func unregisterAppCompletionFn(_ *cobra.Command, _ []string, toComplete string) (
	completion []cobra.Completion,
	shellDirective cobra.ShellCompDirective,
) {
	shellDirective = cobra.ShellCompDirectiveNoFileComp

	workspaceApps, err := workspace.GetApps()
	if err != nil {
		return
	}

	registeredAppNames := []string{}

	for runningApp := range workspaceApps {
		if _, ok := workspaceApps[runningApp]; ok {
			registeredAppNames = append(
				registeredAppNames, common.ShortenAppName(runningApp),
			)
		}
	}

	if len(toComplete) == 0 {
		return registeredAppNames, shellDirective
	}

	for _, rocketApp := range registeredAppNames {
		if toComplete == rocketApp[:len(toComplete)] {
			completion = append(completion, rocketApp)
		}
	}

	return
}
