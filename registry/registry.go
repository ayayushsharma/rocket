package registry

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ayayushsharma/rocket/constants"
)

import (
	"fmt"
	"io"
	"net/http"
	"sync"
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

func FetchRegistryData(urls []string) {
	var wg sync.WaitGroup
	results := make(chan RegistryFetch, len(urls))

	for _, url := range urls {
		wg.Add(1)            // Increment the WaitGroup counter
		go fetchURL(url, &wg, results) // Launch a goroutine for each request
	}

	// Start a separate goroutine to close the results channel once all requests are done
	go func() {
		wg.Wait() // Wait for all fetchURL goroutines to complete
		close(results) // Close the channel to signal no more data will be sent
	}()

	// Read results from the channel
	for result := range results {
		if result.err != nil {
			slog.Debug("Failure in pulling data from registry", "error", result.err)
		}
		fmt.Println(result.data)
	}

	fmt.Println("All requests finished.")
}

