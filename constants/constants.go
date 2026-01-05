package constants

import (
	"log/slog"
	"os"
	"path/filepath"
)

const (
	ApplicationName = "rocket"
	ApplicationPort = 32100
	RouterContainer = "rocket-nginx-router"
)

var (
	NginxConfPath string
	HomePageDir string
	RoutesJson string
	RegisteredAppsJson string
	RegistriesPath string
)

func init() {
	homeDir, err := os.UserHomeDir();
	if err != nil {
		slog.Debug("Could not get user home dir", "error", err)
	}
	configPath := filepath.Join(
		homeDir,
		".config",
		ApplicationName,
	)

	slog.Debug("Default Config Dir", "path", configPath)

	NginxConfPath = filepath.Join(configPath, "nginx/nginx.conf")
	HomePageDir = filepath.Join(configPath, "home-page")
	RoutesJson = filepath.Join(HomePageDir, "static/application.json")
	RegisteredAppsJson = filepath.Join(configPath, "registered.rockets.json")
	RegistriesPath = filepath.Join(configPath, "registries")

	slog.Debug(
		"Default state paths",
		"nginx", NginxConfPath,
		"home", HomePageDir,
		"routes", RoutesJson,
		"registered_apps", RegisteredAppsJson,
		"registries", RegistriesPath,
	)
}
