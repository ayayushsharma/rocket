package cmd

import (
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts Rocket",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Launching... " + constants.ApplicationName)
		var conn containers.ContainerManager
		conn, err := containers.Manager()
		if err != nil {
			slog.Debug("Failed to connect to podman. Exiting")
			os.Exit(1)
		}
		startRouter(conn)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func startRouter(conn containers.ContainerManager) (err error) {
	imageURL := "openresty/openresty:alpine"
	imageExists, err := conn.ImageExists(imageURL)
	if err != nil {
		return
	}

	if !imageExists {
		if err = conn.PullImage(imageURL); err != nil {
			return err
		}
		slog.Debug("Pulled router image")
	}

	networkName := viper.GetString("routes.network")
	slog.Debug("Network found in config", "name", networkName)

	networkExists, err := conn.NetworkExists(networkName);
	if  err != nil {
		return
	}

	if !networkExists {
		if err = conn.CreateNetwork(networkName); err != nil {
			return err
		}
		slog.Debug("Created Network", "name", networkName)
	}

	mountDirs := map[string]string{
		constants.NginxConfPath: "/usr/local/openresty/nginx/conf/nginx.conf",
		constants.HomePageDir:   "/usr/share/nginx/html",
	}

	bindPorts := map[int]int{
		constants.ApplicationPort: 80,
	}

	routerConfig := containers.Config{
		ImageURL:imageURL,
		ContainerName:   constants.RouterContainer,
		ApplicationName: constants.RouterContainer,
		SubDomain:       "app.localhost",
		NetworkName:     networkName,
		MountDirs:       mountDirs,
		BindPorts:       bindPorts,
	}

	err = conn.CreateContainer(routerConfig)
	if err != nil {
		slog.Debug("Could not create router container", "error", err)
		return
	}
	slog.Debug("Created router successfully")

	err = conn.StartService(constants.RouterContainer)
	if err != nil {
		slog.Debug("Could not start router container", "error", err)
		return
	}
	slog.Debug("Router started successfully")

	return nil
}
