// Handlers for local registrations of applications

package register

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry"
)

type routerData struct {
	ContainerURL string
	AppName      string
	Description  string
}

var AppAlreadyRegisteredErr error = errors.New("This app is already registered")
var AppNotRegisteredErr error = errors.New("This app is not registered")
var NoAppSelectedErr error = errors.New("No app selected for registration")

func SelectApplication(apps []registry.AppsOnRegistry) (
	selectedContainer containers.ContainerConfig,
	err error,
) {
	fzfData := []huh.Option[*registry.AppsOnRegistry]{}

	// deduplicating section
	dedup := make(map[string]*registry.AppsOnRegistry)

	for index := range apps {
		fullImageName := common.ImageWithVersion(
			apps[index].App.ImageURL,
			apps[index].App.ImageVersion,
		)
		if _, ok := dedup[fullImageName]; ok {
			if dedup[fullImageName].Priority > apps[index].Priority {
				// override the new image
				dedup[fullImageName] = &apps[index]
				continue
			}
		}
		dedup[fullImageName] = &apps[index]
	}

	// mapping section
	for index := range dedup {
		fzfData = append(fzfData, huh.Option[*registry.AppsOnRegistry]{
			Key: fmt.Sprintf(
				"%-20s %-10s - %s",
				dedup[index].App.ApplicationName,
				dedup[index].App.ImageVersion,
				dedup[index].App.ImageURL,
			),
			Value: dedup[index],
		})
	}

	// selection section
	var selectedAppId *registry.AppsOnRegistry

	err = huh.NewSelect[*registry.AppsOnRegistry]().
		Title("Pick a application").
		Options(fzfData...).
		Value(&selectedAppId).
		Run()

	if err != nil {
		slog.Debug("Failed to select application", "error", err)
		return containers.ContainerConfig{}, err
	}

	return *selectedAppId.App, nil
}

func ReadRegisteredApplications() (
	registeredApps map[string]containers.ContainerConfig,
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

func WriteRegisteredApplications(
	apps map[string]containers.ContainerConfig,
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
	registry, err := ReadRegisteredApplications()
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

func RegisterApplicationToConf(container containers.ContainerConfig) (err error) {
	registry, err := ReadRegisteredApplications()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	if _, exists := registry[container.ContainerName]; exists {
		return AppAlreadyRegisteredErr
	}

	registry[container.ContainerName] = container

	err = WriteRegisteredApplications(registry)
	if err != nil {
		slog.Debug("Failed to write to local register of applications", "error", err)
		return err
	}
	return nil
}

func UnregisterApplicationToConf(containerName string) (err error) {
	registry, err := ReadRegisteredApplications()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	if _, exists := registry[containerName]; !exists {
		return AppNotRegisteredErr
	}

	delete(registry, containerName)

	err = WriteRegisteredApplications(registry)
	if err != nil {
		slog.Debug("Failed to write to local register of applications", "error", err)
		return err
	}
	return nil
}
