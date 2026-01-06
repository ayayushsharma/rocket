package cmd

import (
	"ayayushsharma/rocket/constants"
	// "ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var miscCmd = &cobra.Command{
	Use:   "misc",
	Short: "Launches specified application",

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		// conn, err := containers.Manager()
		// if err != nil {
		//
		// }
		// networks, _ := conn.ListNetworks()
		// for _, val := range networks {
		// 	fmt.Println(val)
		// }

		// conn.CreateContainer(
		// 	containers.ContainerCreateOptions{
		// 		ImageName:       "excalidraw/excalidraw",
		// 		ContainerName:   "test_draw",
		// 		ApplicationName: "test_draw",
		// 		SubDomain:       "draw",
		// 		NetworkName:     "my-network",
		// 		MountDirs:       nil,
		// 		BindPorts:       nil,
		// 		EnvValues:       nil,
		// 		EnvVars:         nil,
		// 	},
		// )
		registries, _ := registry.GetAll()
		fmt.Println(registries)
		registry.FetchRegistries(registries)

	},
}

func init() {
	rootCmd.AddCommand(miscCmd)
}
