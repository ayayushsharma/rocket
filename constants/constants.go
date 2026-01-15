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

var appVersion string

var (
	AppStateDir       string
	UserHomeDir       string
	UserConfigDir     string
	NginxConfPath     string
	HomePageDir       string
	RoutesJson        string
	WorkspaceAppsJson string
	RegistriesPath    string
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Could not get user home dir", "error", err)
		panic(err)
	}

	userConfigPath := os.Getenv("XDG_CONFIG_HOME")
	if userConfigPath == "" {
		userConfigPath = filepath.Join(userHomeDir, ".config")
	}

	rocketConfigDir := filepath.Join(userConfigPath, ApplicationName)
	slog.Debug("Default Config Dir", "path", rocketConfigDir)

	UserHomeDir = userHomeDir
	UserConfigDir = userConfigPath

	AppStateDir = filepath.Join(rocketConfigDir, "state")

	NginxConfPath = filepath.Join(AppStateDir, "nginx/nginx.conf")
	HomePageDir = filepath.Join(AppStateDir, "home-page")
	RoutesJson = filepath.Join(HomePageDir, "static/application.json")

	WorkspaceAppsJson = filepath.Join(rocketConfigDir, "workspace.rockets.json")
	RegistriesPath = filepath.Join(rocketConfigDir, "registries")

	slog.Debug(
		"Default state paths",
		"nginx", NginxConfPath,
		"home", HomePageDir,
		"routes", RoutesJson,
		"registered_apps", WorkspaceAppsJson,
		"registries", RegistriesPath,
	)
}

func GetVersion() string {
	if appVersion == "" {
		return "devel"
	}
	return appVersion
}
