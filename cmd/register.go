package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "registers specified containerised application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Registering... " + constants.ApplicationName)
		conn, err := containers.ConnectPodman()
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			os.Exit(1)
		}

		registries, _ := registry.GetRegistries()
		slog.Debug("Pulled data from registries", "data", registries)
		data := registry.FetchRegistryData(registries)

		appToRegister, err := registry.ApplicationToRegister(data)
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			os.Exit(1)
		}

		networkName := viper.GetString("routes.network")
		slog.Debug("Network found", "name", networkName)
		appToRegister.NetworkName = networkName
		slog.Debug("App Data", "data", appToRegister)

		image := appToRegister.ImageURL
		if strings.Trim(appToRegister.ImageVersion, " ") != "" {
			image = fmt.Sprintf(
				"%s:%s", image, strings.Trim(appToRegister.ImageVersion, " "),
			)
		}
		err = conn.PullImage(image)
		if err != nil {
			slog.Debug("Pulling image failed for application container", "error", err)
			os.Exit(1)
		}

		_, err = registry.RegisterApplicationToConf(appToRegister)
		if err != nil {
			slog.Debug(
				"Failed to register application container to configuration",
				"error",
				err,
			)
			os.Exit(1)
		}

		_, err = registry.RefreshRouterConf()
		if err != nil {
			slog.Debug("Failed to register application to routes", "error", err)
			os.Exit(1)
		}

		err = conn.CreateContainer(appToRegister)
		if err != nil {
			slog.Debug("Failed to register application container", "error", err)
			os.Exit(1)
		}

		err = conn.StartService(appToRegister.ContainerName)
		if err != nil {
			slog.Debug("Failed to start application", "error", err)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
