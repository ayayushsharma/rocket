// Handlers for local registrations of applications

package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	// "ayayushsharma/rocket/common"
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
)

type routerData struct {
	ContainerURL string
	AppName      string
	Description  string
}

var AppAlreadyRegisteredErr error = errors.New("This app is already registered")
var AppNotRegisteredErr error = errors.New("This app is not registered")
var NoAppSelectedErr error = errors.New("No app selected for registration")

func GetRegistered() (
	registeredApps map[string]containers.Config,
	err error,
) {
	data, err := os.ReadFile(constants.RegisteredAppsJson)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		data = []byte("{}")
	}

	if err := json.Unmarshal(data, &registeredApps); err != nil {
		slog.Debug("Registry Unmarshalling failed", "error", err)
		return nil, err
	}

	return registeredApps, nil
}

func WriteRegisteredApps(
	apps map[string]containers.Config,
) (err error) {
	jsonData, err := json.MarshalIndent(apps, "", "  ")
	if err != nil {
		slog.Debug("Error marshalling registered apps JSON", "error", err)
		return
	}

	err = os.WriteFile(constants.RegisteredAppsJson, jsonData, 0644)
	if err != nil {
		slog.Debug("Error writing to registered apps JSON", "error", err)
		return
	}

	slog.Debug("Successfully wrote registered app conf")
	return nil
}

func RefreshRouterConf() (err error) {
	registry, err := GetRegistered()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	routes := map[string]routerData{}

	for _, val := range registry {
		redirectionPort := fmt.Sprintf(
			"http://%s:%d",
			val.ContainerName,
			val.ExposeHttpPort,
		)
		routes[val.SubDomain] = routerData{
			ContainerURL: redirectionPort,
			AppName:      val.ApplicationName,
		}
	}

	jsonData, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		slog.Debug("Error marshalling route JSON", "error", err)
		return err
	}

	err = os.WriteFile(constants.RoutesJson, jsonData, 0644)
	if err != nil {
		slog.Debug("Error writing to routes JSON", "error", err)
		return err
	}

	slog.Debug("Successfully wrote routes conf")
	return nil

}

func RegisterApplicationToConf(container containers.Config) (err error) {
	registry, err := GetRegistered()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	if _, exists := registry[container.ContainerName]; exists {
		return AppAlreadyRegisteredErr
	}

	registry[container.ContainerName] = container

	err = WriteRegisteredApps(registry)
	if err != nil {
		slog.Debug("Failed to write to local register of applications", "error", err)
		return err
	}
	return nil
}

func UnregisterApplicationToConf(containerName string) (err error) {
	registry, err := GetRegistered()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	if _, exists := registry[containerName]; !exists {
		return AppNotRegisteredErr
	}

	delete(registry, containerName)

	err = WriteRegisteredApps(registry)
	if err != nil {
		slog.Debug("Failed to write to local register of applications", "error", err)
		return err
	}
	return nil
}
