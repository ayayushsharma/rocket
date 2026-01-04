package registry

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"io"
	"net/http"
	"encoding/json"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry/schema"

	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	fzf "github.com/junegunn/fzf/src"
)

func GetRegistries() (registries []string, err error){
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Debug("Could not get home dir location", "error", err)
		return nil, err
	}
	configFile := filepath.Join(
		userHomeDir,
		".config",
		constants.ApplicationName,
		"registries",
	)
	data, err := os.ReadFile(configFile)
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

type RegistryFetch struct {
	data string
	err error
}

func fetchURL(url string, wg *sync.WaitGroup, results chan<- RegistryFetch) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		results <- RegistryFetch{"", fmt.Errorf("Error fetching %s: %v", url, err)}
		return
	}
	defer resp.Body.Close()

	// Read the response body (necessary for connection reuse in some cases)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- RegistryFetch{"", fmt.Errorf("Error reading body from %s: %v", url, err)}
		return
	}

	results <- RegistryFetch{string(data), nil}
}

func FetchRegistryData(
	urls []string,
) (containersData []containers.ContainerConfig){
	var wg sync.WaitGroup
	results := make(chan RegistryFetch, len(urls))

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
		fmt.Println(result.data)
		registryData, err := schema.InterpreterV1(result.data)
		if err != nil {
			slog.Debug("Parsing data from registry failed", "error", err)
			continue
		}
		containersData = append(containersData, registryData...)
	}

	fmt.Println("All requests finished.")
	return containersData
}


func ApplicationToRegister(
    containerData []containers.ContainerConfig,
) (selectedContainer containers.ContainerConfig, err error) {
    mapping := make(map[string]containers.ContainerConfig, len(containerData))
    fzfData := make([]string, 0, len(containerData))

    for _, c := range containerData {
        id := uuid.NewString()
        mapping[id] = c
        fzfData = append(fzfData, fmt.Sprintf("%s\t%s", c.ApplicationName, id))
    }

    inputChan := make(chan string)
    outputChan := make(chan string)

    var wg sync.WaitGroup

    wg.Go(func() {
        for _, s := range fzfData {
            inputChan <- s
        }
        close(inputChan)
    })

    var selectedIDs []string
    wg.Go(func() {
        for line := range outputChan {
            parts := strings.Split(line, "\t")
            if len(parts) > 0 {
                selectedIDs = append(selectedIDs, parts[len(parts)-1])
            }
        }
    })

    options, err := fzf.ParseOptions(
        true,
        []string{
            "--reverse",
            "--border",
            "--height=40%",
            "--with-nth=1",
            "--delimiter=\t",
        },
    )
    if err != nil {
        return containers.ContainerConfig{}, err
    }

    options.Input = inputChan
    options.Output = outputChan

	exit, err := fzf.Run(options)

    close(outputChan)

    wg.Wait()

	if exit == fzf.ExitNoMatch || 
		exit == fzf.ExitError ||
		exit == fzf.ExitInterrupt {
		return containers.ContainerConfig{}, errors.New("Failed to get app ID")	
	}

    if err != nil || len(selectedIDs) == 0 {
        return containers.ContainerConfig{}, err
    }

    if selected, ok := mapping[selectedIDs[0]]; ok && exit == fzf.ExitOk{
        return selected, nil
    }
    return containers.ContainerConfig{}, errors.New("Unknown value selected")
}

var AppAlreadyRegistered error = errors.New("This app is already registered")


func ReadRegisteredApplications() (
	registry map[string]containers.ContainerConfig, 
	err error,
) {
	data, err := os.ReadFile(constants.RegisteredAppsJson)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		data = []byte("{}")
	}

	if err := json.Unmarshal(data, &registry); err != nil {
		slog.Debug("Registry Unmarshalling failed", "error" ,err)
		return nil, err
	}

	return registry, nil
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

func RegisterApplicationToConf(
	container containers.ContainerConfig,
) (err error) {
	registry, err := ReadRegisteredApplications()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return err
	}

	if _, exists := registry[container.ContainerName]; exists {
		return AppAlreadyRegistered
	}

	registry[container.ContainerName] = container

	err = WriteRegisteredApplications(registry)
	if err != nil {
		slog.Debug("Failed to write to local register of applications", "error", err)
		return err
	}
	return nil
}

type RouterData struct {
	ContainerURL string
	AppName string
	Description string
}

func RefreshRouterConf() (ok bool, err error) {
	registry, err := ReadRegisteredApplications()
	if err != nil {
		slog.Debug("Failed to read locally registered applications", "error", err)
		return false, err
	}

	routes := map[string]RouterData{}

	for _, val := range registry {
		redirectionPort := fmt.Sprintf(
			"http://%s:%d",
			val.ContainerName,
			val.ExposeHttpPort,
		)
		routes[val.SubDomain] = RouterData{
			ContainerURL: redirectionPort,
			AppName: val.ApplicationName,
		}
	}

	jsonData, err := json.MarshalIndent(routes, "", "  ")
	if err != nil {
		slog.Debug("Error marshalling route JSON", "error", err)
		return false, err
	}

	err = os.WriteFile(constants.RoutesJson, jsonData, 0644)
	if err != nil {
		slog.Debug("Error writing to routes JSON", "error", err)
		return false, err
	}

	slog.Debug("Successfully wrote routes conf")
	return true, nil

}
