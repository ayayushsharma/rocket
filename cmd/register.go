package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/register"
	"ayayushsharma/rocket/registry"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "registers specified containerised application",

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		slog.Debug("Registering... " + constants.ApplicationName)
		conn, err := containers.ConnectPodman()
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			return
		}

		registries, err := registry.GetAll()

		if err != nil {
			slog.Debug("Failed to pull list of registries", "error", err)
			return
		}

		slog.Debug("Pulled data from registries", "data", registries)
		data := registry.FetchRegistries(registries)

		appToRegister, err := register.SelectApplication(data)
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			return
		}

		return

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
			return
		}

		err = register.RegisterApplicationToConf(appToRegister)
		if err != nil {
			slog.Debug(
				"Failed to register application container to configuration",
				"error",
				err,
			)
			return
		}

		err = register.RefreshRouterConf()
		if err != nil {
			slog.Debug("Failed to register application to routes", "error", err)
			return
		}

		err = conn.CreateContainer(appToRegister)
		if err != nil {
			slog.Debug("Failed to register application container", "error", err)
			return
		}

		err = conn.StartService(appToRegister.ContainerName)
		if err != nil {
			slog.Debug("Failed to start application", "error", err)
			return
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
