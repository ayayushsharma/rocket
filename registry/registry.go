package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/charmbracelet/huh"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry/schema"
)


type registryFetch struct {
	data string
	err error
}

type routerData struct {
	ContainerURL string
	AppName string
	Description string
}

var AppAlreadyRegisteredErr error = errors.New("This app is already registered")
var AppNotRegisteredErr error = errors.New("This app is not registered")
var NoAppSelectedErr error = errors.New("No app selected for registration")


func GetRegistries() (registries []string, err error){
	registriesFile := constants.RegistriesPath

	data, err := os.ReadFile(registriesFile)
	if err != nil {
		slog.Debug("Could not read registry file", "error", err)
		return nil, err
	}
	
	registryLines := strings.Split(string(data),"\n")

	for _, line := range(registryLines) {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			// these are comments
			continue
		}
		registries = append(registries, line)
	}

	return registries, nil
}


func fetchURL(url string, wg *sync.WaitGroup, results chan<- registryFetch) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		results <- registryFetch{"", fmt.Errorf("Error fetching %s: %v", url, err)}
		return
	}
	defer resp.Body.Close()

	// Read the response body (necessary for connection reuse in some cases)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- registryFetch{"", fmt.Errorf("Error reading body from %s: %v", url, err)}
		return
	}

	results <- registryFetch{string(data), nil}
}


func FetchRegistryData(urls []string) (containerCfgs []containers.ContainerConfig){
	var wg sync.WaitGroup
	results := make(chan registryFetch, len(urls))

	for _, url := range urls {
		wg.Add(1)
		go fetchURL(url, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			slog.Debug("Failure in pulling data from registry", "error", result.err)
		}
		registryData, err := schema.InterpreterV1(result.data)
		if err != nil {
			slog.Debug("Parsing data from registry failed", "error", err)
			continue
		}
		containerCfgs = append(containerCfgs, registryData...)
	}

	fmt.Println("All requests finished.")
	return containerCfgs
}


func ApplicationToRegister(
    containerCfgs []containers.ContainerConfig,
) (selectedContainer containers.ContainerConfig, err error) {
    mapping := make(map[string]containers.ContainerConfig, len(containerCfgs))
    fzfData := []huh.Option[string]{}

    for _, c := range containerCfgs {
        id := uuid.NewString()
        mapping[id] = c
        fzfData = append(fzfData, huh.Option[string]{
			Key: c.ApplicationName,
			Value: id,
		})
    }

	var selectedAppId string

	err = huh.NewSelect[string]().
	Title("Pick a application").
	Options(fzfData...).
	Value(&selectedAppId).
	Run()

	if err != nil {
		slog.Debug("Failed to select application", "error", err)
		return containers.ContainerConfig{}, err
	}

	if selected, ok := mapping[selectedAppId]; ok {
        return selected, nil
    }
    return containers.ContainerConfig{}, errors.New("Unknown value selected")
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
		slog.Debug("Registry Unmarshalling failed", "error" ,err)
		return nil, err
	}

	return registeredApps, nil
}

func WriteRegisteredApplications(
	apps map[string]containers.ContainerConfig,
) (err error){
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
			AppName: val.ApplicationName,
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
