// Managers for local workspace configuration of applications

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

func getWorkspace() (workspace workspaceSchema, err error) {
	data, err := os.ReadFile(constants.WorkspaceAppsJson)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return
		}
	}

	if data != nil {
		if err = json.Unmarshal(data, &workspace); err != nil {
			return
		}
		return
	}

	workspace = workspaceSchema{
		Applications: map[string]containers.Config{},
	}

	return workspace, nil
}

func updateWorkspace(apps map[string]containers.Config) (err error) {
	workspace, err := getWorkspace()
	if err != nil {
		return
	}

	workspace.Applications = apps

	jsonData, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return
	}

	err = os.WriteFile(constants.WorkspaceAppsJson, jsonData, 0644)
	if err != nil {
		return
	}

	slog.Debug("Successfully wrote registered app conf")
	return nil
}

func GetApps() (
	workspaceApps map[string]containers.Config,
	err error,
) {
	workspace, err := getWorkspace()
	if err != nil {
		return
	}
	return workspace.Applications, nil
}

func SyncRouter() (err error) {
	registry, err := GetApps()
	if err != nil {
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
		return err
	}

	err = os.WriteFile(constants.RoutesJson, jsonData, 0644)
	if err != nil {
		return err
	}

	slog.Debug("Successfully wrote routes conf")
	return nil

}

func Register(container containers.Config) (err error) {
	apps, err := GetApps()
	if err != nil {
		return
	}

	if _, exists := apps[container.ContainerName]; exists {
		return &AppAlreadyRegisteredErr{
			ContainerName: container.ContainerName,
		}
	}

	apps[container.ContainerName] = container

	err = updateWorkspace(apps)
	if err != nil {
		return err
	}

	err = SyncRouter()
	if err != nil {
		return
	}

	return nil
}

func Unregister(containerName string) (err error) {
	registry, err := GetApps()
	if err != nil {
		return err
	}

	if _, exists := registry[containerName]; !exists {
		return AppNotRegisteredErr
	}

	delete(registry, containerName)

	err = updateWorkspace(registry)
	if err != nil {
		return err
	}

	err = SyncRouter()
	if err != nil {
		return
	}

	return nil
}

func GetAppCfg(appName string) (app containers.Config, err error) {
	allApps, err := GetApps()
	if err != nil {
		return
	}

	app, ok := allApps[appName]

	if !ok {
		return app, AppNotRegisteredErr
	}

	return app, nil
}
