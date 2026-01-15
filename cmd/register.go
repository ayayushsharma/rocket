package cmd

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/registry"
	"ayayushsharma/rocket/workspace"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Registers containerised application from selection menu",

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		slog.Debug("Registering... " + constants.ApplicationName)
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

		appToRegister, err := registry.SelectApplication(data)
		if err != nil {
			slog.Debug("Failed to select application", "error", err)
			return
		}

		networkName := viper.GetString("routes.network")
		slog.Debug("Network found", "name", networkName)
		appToRegister.NetworkName = networkName
		err = workspace.Register(appToRegister)

		if err != nil {
			var alreadyRegistered *workspace.AppAlreadyRegisteredErr
			if errors.As(err, &alreadyRegistered) {
				fmt.Printf(
					"Already registered as '%s' \n",
					alreadyRegistered.ContainerName,
				)
				fmt.Println("Edit it's conf and sync to get desired app state")
				return nil
			}
			slog.Debug("Failed to register app to workspace", "error", err)
			return
		}

		fmt.Println("Application Successfully registered as:")
		fmt.Println(appToRegister.ContainerName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
