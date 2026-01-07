package cmd

import (
	"fmt"
	"log/slog"
	"os"
	// "sync"

	"github.com/spf13/cobra"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/workspace"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launches specified application",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// launchAll := false
		if len(args) > 0 {
			// var wg sync.WaitGroup
			for _, appName := range args {
				// wg.Go(func() {
				launchApp(appName)
				// })
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}

func launchApp(appName string) (err error) {
	slog.Debug("Launching... " + appName)
	conn, err := containers.Manager()
	if err != nil {
		slog.Debug("Failed to connect to podman. Exiting")
		return
	}

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
