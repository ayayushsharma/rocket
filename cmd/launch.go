package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/workspace"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launches specified application",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			return
		}

		isLaunchAll := viper.GetBool("all")
		if isLaunchAll {
			launchAll(conn)
			return
		}

		if len(args) > 0 {
			for _, appName := range args {
				launchApp(conn, common.CompleteAppName(appName))
			}
		}

		return nil
	},
	ValidArgsFunction: launchAppCompletionFn,
}

func init() {
	rootCmd.AddCommand(launchCmd)
	launchCmd.Flags().Bool("all", false, "Launch all the registered apps")
}

func launchApp(conn containers.ContainerManager, appName string) (err error) {
	slog.Debug("Launching... " + appName)

	exists, err := conn.ContainerExists(appName)

	if !exists {
		if err = createApp(conn, appName); err != nil {
			slog.Debug("Failed to create app", "error", err)
			return err
		}
	}

	err = conn.StartService(appName)
	if err != nil {
		slog.Debug("Failed to start container. Exiting")
		os.Exit(1)
	}
	slog.Debug("Successfully started application", "application", appName)

	return nil
}

func createApp(conn containers.ContainerManager, appName string) (err error) {
	appCfg, err := workspace.GetAppCfg(appName)
	if err != nil {
		if err == workspace.AppNotRegisteredErr {
			fmt.Printf("App not registered: %s", appName)
		}
		return err
	}

	image := common.ImageWithVersion(appCfg.ImageURL, appCfg.ImageVersion)
	exists, err := conn.ImageExists(image)

	if err != nil {
		slog.Debug("App image could not be checked if it exists", "error", err)
		return err
	}

	if !exists {
		slog.Debug("Image name", "image", image)
		err := conn.PullImage(image)
		if err != nil {
			slog.Debug("App image could not be pulled", "error", err)
			return err
		}
	}

	err = conn.CreateContainer(appCfg)
	if err != nil {
		slog.Error("Failed to create App container", "error", err)
		return err
	}

	return nil
}

func launchAll(conn containers.ContainerManager) (err error) {
	apps, err := workspace.GetApps()
	if err != nil {
		return err
	}

	var storeErr error
	for appName := range apps {
		err = launchApp(conn, appName)
		if err != nil {
			storeErr = err
		}
	}

	return storeErr
}

func launchAppCompletionFn(_ *cobra.Command, _ []string, toComplete string) (
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

	for workspaceApp := range workspaceApps {
		if !slices.Contains(runningApps, workspaceApp) {
			runningRockets = append(
				runningRockets, common.ShortenAppName(workspaceApp),
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
