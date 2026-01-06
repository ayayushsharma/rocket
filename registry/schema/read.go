package schema

import (
	"encoding/json"
	"log/slog"

	"ayayushsharma/rocket/containers"
)

type registrySchema struct {
	Version int `json:"version"`
}

type registryReader func(
	registryData string,
) (
	parsedData []containers.Config,
	err error,
)

var readerMap map[int]registryReader

// Parses supplied registry data.
// Handles registry versions by itself
func Parse(
	registryData string,
) (parsedData []containers.Config, err error) {
	var registry registrySchema
	if err := json.Unmarshal([]byte(registryData), &registry); err != nil {
		slog.Debug("Failed to get registry version", "error", err)
		return nil, err
	}

	versionedReader := readerMap[registry.Version]
	return versionedReader(registryData)
}

func init() {
	readerMap = make(map[int]registryReader)
	readerMap[1] = parseV1Registry
}
