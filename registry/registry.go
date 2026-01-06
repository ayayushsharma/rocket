package registry

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"

	"ayayushsharma/rocket/common"
	"ayayushsharma/rocket/constants"
	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/registry/schema"
)

type registryPull struct {
	rank int // lower number means higher priority
	data string
	err  error
}

type appFromRegistry struct {
	rank int // lower number means higher priority
	app  *containers.Config
}

// Returns list of all registries present on the local registry config
func GetAll() (registries []string, err error) {
	registriesFile := constants.RegistriesPath

	data, err := os.ReadFile(registriesFile)
	if err != nil {
		slog.Debug("Could not read registry file", "error", err)
		return nil, err
	}

	registryLines := strings.SplitSeq(string(data), "\n")

	for line := range registryLines {
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
func fetchOverHTTP(url string) (data []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("Error fetching %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("Error reading body from %s: %v", url, err)
		return
	}

	return data, nil
}

// fetch registry over the internet
func fetchOverDisk(path string) (data []byte, err error) {
	data, err = os.ReadFile(path)
	if err != nil {
		slog.Debug("Could not read registry file", "error", err)
		err = fmt.Errorf("Could not read registry from %s: %v", path, err)
		return
	}
	return
}

// Fetches registry data over any types of registry type
// - HTTP type registry
// - local file type registry
func fetch(
	registryPriority int,
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
		data, err := fetchOverHTTP(registryURI)
		results <- registryPull{
			rank: registryPriority,
			data: string(data),
			err:  err,
		}
		return
	}

	_, err = os.Stat(registryURI)
	isDisk := err == nil
	if isDisk {
		data, err := fetchOverDisk(registryURI)
		results <- registryPull{
			rank: registryPriority,
			data: string(data),
			err:  err,
		}
		return
	}

	results <- registryPull{
		rank: registryPriority,
		data: "",
		err:  err,
	}
	slog.Debug("No a valid registry path", "error", registryURI)
}

// fetches all registries for available applications
func FetchRegistries(registryURIs []string) (
	containerCfgs []appFromRegistry,
) {
	var wg sync.WaitGroup
	results := make(chan registryPull, len(registryURIs))

	for priority, uri := range registryURIs {
		wg.Add(1)
		go fetch(priority, uri, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			slog.Debug("Failure in pulling data from registry", "error", result.err)
			continue
		}
		registryData, err := schema.Parse(result.data)
		if err != nil {
			slog.Debug("Parsing data from registry failed", "error", err)
			continue
		}
		appsWithPriority := []appFromRegistry{}
		for index := range registryData {
			appsWithPriority = append(appsWithPriority, appFromRegistry{
				rank: result.rank,
				app:  &registryData[index],
			})
		}

		containerCfgs = append(containerCfgs, appsWithPriority...)
	}

	slog.Debug("Finished data pull from all registry")

	return containerCfgs
}

func SelectApplication(apps []appFromRegistry) (
	selected containers.Config,
	err error,
) {
	fzfData := []huh.Option[*appFromRegistry]{}

	// deduplicating section
	dedup := make(map[string]*appFromRegistry)

	for index := range apps {
		fullImageName := common.ImageWithVersion(
			apps[index].app.ImageURL,
			apps[index].app.ImageVersion,
		)
		if _, ok := dedup[fullImageName]; ok {
			if dedup[fullImageName].rank > apps[index].rank {
				// override the new image
				dedup[fullImageName] = &apps[index]
				continue
			}
		}
		dedup[fullImageName] = &apps[index]
	}

	// mapping section
	for index := range dedup {
		fzfData = append(fzfData, huh.Option[*appFromRegistry]{
			Key: fmt.Sprintf(
				"%-20s %-10s - %s",
				dedup[index].app.ApplicationName,
				dedup[index].app.ImageVersion,
				dedup[index].app.ImageURL,
			),
			Value: dedup[index],
		})
	}

	// selection section
	var selectedAppId *appFromRegistry

	err = huh.NewSelect[*appFromRegistry]().
		Title("Pick a application").
		Options(fzfData...).
		Value(&selectedAppId).
		Run()

	if err != nil {
		slog.Debug("Failed to select application", "error", err)
		return containers.Config{}, err
	}

	return *selectedAppId.app, nil
}
