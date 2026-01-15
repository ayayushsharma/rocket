package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/workspace"
)

var stopCmd = &cobra.Command{
	Use:   "stop [container_name]",
	Short: "Stops rocket applications",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var conn containers.ContainerManager
		conn, err = containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			return
		}
		if len(args) > 0 {
			for _, appName := range args {
				stopApp(conn, common.CompleteAppName(appName))
			}
		}

		return nil
	},
	ValidArgsFunction: stopAppCompletionFn,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stopApp(conn containers.ContainerManager, appName string) (err error) {
	slog.Debug("Stopping... " + constants.ApplicationName)
	err = conn.StopService(appName)
	if err != nil {
		slog.Debug("Failed to start container. Exiting")
		return
	}
	slog.Debug("Successfully stopped application", "application", appName)

	return nil
}

func stopAppCompletionFn(cmd *cobra.Command, args []string, toComplete string) (
	completion []cobra.Completion,
	shellDirective cobra.ShellCompDirective,
) {
	shellDirective = cobra.ShellCompDirectiveNoFileComp
	runningRockets := []string{}
	var conn containers.ContainerManager
	conn, err := containers.Manager()
	if err != nil {
		return
	}

	runningApps, err := conn.ListContainers()
	if err != nil {
		return
	}

	workspaceApps, err := workspace.GetApps()
	if err != nil {
		return
	}

	for _, runningApp := range runningApps {
		if _, ok := workspaceApps[runningApp]; ok {
			runningRockets = append(
				runningRockets, common.ShortenAppName(runningApp),
			)
		}
	}

	if len(toComplete) == 0 {
		return runningRockets, shellDirective
	}

	for _, rocketApp := range runningRockets {
		if toComplete == rocketApp[:len(toComplete)] {
			completion = append(completion, rocketApp)
		}
	}

	return
}
