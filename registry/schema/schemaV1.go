package schema

import (
	"encoding/json"
	"log/slog"
	"strings"

	"ayayushsharma/rocket/containers"
	"ayayushsharma/rocket/common"
)

type RegistryApplicationsV1 struct {
	Name 				string 	`json:"name"`
	ArtifactoryUrl 		string 	`json:"artifactoryUrl"`
	Version 			string 	`json:"version"`
	HttpPort 			int 	`json:"httpPort"`
	Hostname 			string 	`json:"hostname"`
}

type RegistryDataV1 struct {
	Version 			string 						`json:"version"`
	Application 		[]RegistryApplicationsV1 	`json:"applications"`
}

func InterpreterV1(
	registryData string,
) (parsedData []containers.ContainerConfig, err error) {
	var registry RegistryDataV1
	if err := json.Unmarshal([]byte(registryData), &registry); err != nil {
		slog.Debug("Registry Unmarshalling failed", "error" ,err)
		return nil, err
	}
	for _, app := range registry.Application {
		containerName := common.CreateContainerName(
			app.Name,
			app.ArtifactoryUrl,
			app.Version,
		)
		hostName := app.Hostname
		endsWithLocalhost := strings.HasSuffix(hostName, ".localhost")
		if !endsWithLocalhost {
			hostName = hostName + ".localhost"
		}

		application := containers.ContainerConfig{
			ApplicationName: app.Name,
			ContainerName: containerName,
			ImageURL: app.ArtifactoryUrl, 
			ImageVersion: app.Version,
			SubDomain: hostName,
			ExposeHttpPort: app.HttpPort,
		}
		parsedData = append(parsedData, application)		
	}

	return parsedData, nil
}
