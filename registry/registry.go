package registry

import (
	"fmt"
	"io"
	"net/url"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry/schema"
)


type registryPull struct {
	data string
	err error
}

// Returns list of all registries present on the local registry config
func GetAll() (registries []string, err error){
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


// fetch registry over the internet
func fetchOverHTTP(
	url string,
	wg *sync.WaitGroup,
	results chan<- registryPull,
) {
	resp, err := http.Get(url)
	if err != nil {
		results <- registryPull{
			"", fmt.Errorf("Error fetching %s: %v", url, err),
		}
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- registryPull{
			"", fmt.Errorf("Error reading body from %s: %v", url, err),
		}
		return
	}

	results <- registryPull{string(data), nil}
}

// fetch registry over the internet
func fetchOverDisk(
	path string,
	wg *sync.WaitGroup,
	results chan<- registryPull,
) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("Could not read registry file", "error", err)
		results <- registryPull{
			"", fmt.Errorf("Could not read registry from %s: %v", path, err),
		}
	}
	results <- registryPull{string(data), nil}
}


// Fetches registry data over any types of registry type
// - HTTP type registry
// - local file type registry
func fetch(
	registryURI string,
	wg *sync.WaitGroup,
	results chan<- registryPull,
) {
	defer wg.Done()
	u, err := url.Parse(registryURI)
	isURL := err == nil &&
		u.Scheme != "" &&
		u.Host != "" &&
		(strings.EqualFold(u.Scheme, "http") ||
		strings.EqualFold(u.Scheme, "https"))

	if isURL {
		fetchOverHTTP(registryURI, wg, results)
		return
	}

    _, err = os.Stat(registryURI)
	isDisk := err == nil 
	if isDisk {
		fetchOverDisk(registryURI, wg, results)
		return
	}

	slog.Debug("No a valid registry path", "error", registryURI)
}

// fetches all registries for available applications
func FetchRegistries(registryURIs []string) (
	containerCfgs []containers.ContainerConfig,
) {
	var wg sync.WaitGroup
	results := make(chan registryPull, len(registryURIs))

	for _, uri := range registryURIs {
		wg.Add(1)
		go fetch(uri, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			slog.Debug("Failure in pulling data from registry", "error", result.err)
		}
		registryData, err := schema.Parse(result.data)
		if err != nil {
			slog.Debug("Parsing data from registry failed", "error", err)
			continue
		}
		containerCfgs = append(containerCfgs, registryData...)
	}

	slog.Debug("Finished data pull from all registry")

	return containerCfgs
}
