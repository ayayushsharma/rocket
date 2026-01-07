package containers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/podman/v6/pkg/bindings"
	"github.com/containers/podman/v6/pkg/bindings/containers"
	"github.com/containers/podman/v6/pkg/bindings/images"
	"github.com/containers/podman/v6/pkg/bindings/network"
	"github.com/containers/podman/v6/pkg/specgen"

	spec "github.com/opencontainers/runtime-spec/specs-go"
	nettypes "go.podman.io/common/libnetwork/types"
)

type connectionFileV2 struct {
	Connection struct {
		Default     string                       `json:"Default"`
		Connections map[string]connectionEntryV2 `json:"Connections"`
	} `json:"Connection"`
	// "Farm" may exist, we ignore it.
}

type connectionEntryV2 struct {
	URI      string `json:"URI"`
	Identity string `json:"Identity"`
	// Some builds add flags like IsMachine, TLS*, etc.
}

// Legacy schema (array of objects produced by `podman system connection list --format=json`)
type connectionEntryV1 struct {
	Name      string `json:"Name"`
	URI       string `json:"URI"`
	Identity  string `json:"Identity"`
	Default   bool   `json:"Default"`
	ReadWrite bool   `json:"ReadWrite"`
}

func connectionsPath() (string, error) {
	if p := os.Getenv("PODMAN_CONNECTIONS_CONF"); p != "" {
		return p, nil
	}

	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	var base string
	if xdgHome != "" {
		base = xdgHome
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "containers", "podman-connections.json"), nil
}

// loadDefaultConnection reads the file and returns (uri, identity).
func loadDefaultConnection() (string, string, error) {
	path, err := connectionsPath()
	if err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("open %s: %w", path, err)
	}

	// Try newer V2 schema (object with Connection.Default + map)
	var v2 connectionFileV2
	if err2 := json.Unmarshal(data, &v2); err2 == nil && v2.Connection.Connections != nil {
		def := v2.Connection.Default
		if def == "" {
			return "", "", errors.New("connections file has empty Default")
		}
		entry, ok := v2.Connection.Connections[def]
		if !ok {
			return "", "", fmt.Errorf("default connection %q not found in Connections", def)
		}
		if entry.URI == "" {
			return "", "", errors.New("default connection has empty URI")
		}
		return entry.URI, entry.Identity, nil
	}

	// Fallback: legacy V1 schema (array)
	var v1 []connectionEntryV1
	if err1 := json.Unmarshal(data, &v1); err1 == nil && len(v1) > 0 {
		for _, c := range v1 {
			if c.Default {
				if c.URI == "" {
					return "", "", errors.New("default connection (v1) has empty URI")
				}
				return c.URI, c.Identity, nil
			}
		}
		return "", "", errors.New("no default connection in legacy array")
	}

	return "", "", errors.New("unrecognized podman connections schema")
}

func defaultSocketURI() (string, error) {
	uri, _, err := loadDefaultConnection()
	return uri, err
}

type PodManContext struct {
	context.Context
}

func connectPodman() (PodManContext, error) {
	socketURI, err := defaultSocketURI()
	if err != nil {
		slog.Error("Couldn't connect to Podman", "error", err)
		slog.Error("Check if Podman service is running on the machine")
		return PodManContext{nil}, err
	}

	slog.Debug("Socket URI found", "uri", socketURI)
	conn, err := bindings.NewConnection(context.Background(), socketURI)

	if err != nil {
		slog.Error("Couldn't connect to podman service", "error", err)
		return PodManContext{nil}, err
	}

	slog.Debug("Connected to podman")
	return PodManContext{conn}, nil
}

func (conn PodManContext) PullImage(imageName string) error {
	_, err := images.Pull(conn, imageName, nil)
	if err != nil {
		slog.Debug("Failed to pull container image", "error", err)
		return err
	}

	slog.Debug("Pulled Podman image", "name", imageName)
	return nil
}

func (conn PodManContext) RemoveImage(imageName string) error {
	imageList := []string{imageName}
	report, errs := images.Remove(conn, imageList, nil)
	if errs != nil {
		slog.Debug("Failed to remove images")
		return errs[0]
	}
	slog.Info("Image removal report", "report", report)

	return nil
}

func (conn PodManContext) ImageExists(imageName string) (exists bool, err error) {
	exists, err = images.Exists(conn, imageName, nil)
	if err != nil {
		slog.Debug("Failed to check if container image exists", "error", err)
		return false, err
	}
	return exists, nil
}

// func (conn PodManContext) ListContainers() ([]string, error) {
// 	containerList, err := containers.List(conn, nil);
// 	if  err != nil {
// 		return
// 	}
//
// 	return
// }

func (conn PodManContext) ContainerExists(containerName string) (exists bool, err error) {
	exists, err = containers.Exists(conn, containerName, nil)
	return
}

func (conn PodManContext) CreateContainer(options Config) (err error) {
	image := options.ImageURL

	if strings.Trim(options.ImageVersion, " ") != "" {
		image = options.ImageURL + ":" + strings.Trim(options.ImageVersion, " ")
	}

	s := specgen.NewSpecGenerator(image, false)
	s.Name = options.ContainerName
	s.Hostname = options.ContainerName

	if s.Labels == nil {
		s.Labels = map[string]string{}
	}
	s.Labels["app.name"] = options.ApplicationName
	s.Labels["app.container"] = options.ContainerName
	s.Labels["app.subdomain"] = options.SubDomain

	if options.NetworkName != "" {
		if s.Networks == nil {
			s.Networks = map[string]nettypes.PerNetworkOptions{}
		}
		s.Networks[options.NetworkName] = nettypes.PerNetworkOptions{}
	}

	for hostPort, containerPort := range options.BindPorts {
		s.PortMappings = append(s.PortMappings, nettypes.PortMapping{
			HostPort:      uint16(hostPort),
			ContainerPort: uint16(containerPort),
			Protocol:      "tcp",
		})
	}

	// Mounts: host path -> container path as bind mounts
	for hostDir, containerDir := range options.MountDirs {
		s.Mounts = append(s.Mounts, spec.Mount{
			Source:      hostDir,
			Destination: containerDir,
			Type:        "bind",
			// "rbind" preserves sub-mount propagation
			// "rw" for read-write / Use "ro" for read-only
			Options: []string{"rbind", "ro"},
		})
	}

	s.RestartPolicy = "always"

	// s.Env = map[string]string{
	//     "APP_NAME":  applicationName,
	//     "SUBDOMAIN": subdomain,
	// }

	containerExists, err := containers.Exists(conn, options.ContainerName, nil)
	if err != nil {
		slog.Debug("Failed to check if container already exists", "error", err)
		return err
	}
	if containerExists {
		slog.Debug("Container already exists")
		return nil
	}

	ctr, err := containers.CreateWithSpec(conn, s, nil)
	if err != nil {
		return fmt.Errorf("create container %q failed: %w", options.ContainerName, err)
	}
	slog.Debug("Container created", "response", ctr)
	return nil
}

func (conn PodManContext) RemoveContainer(containerName string, force bool) (err error) {
	options := containers.RemoveOptions{
		Force: &force,
	}
	_, err = containers.Remove(conn, containerName, &options)
	return
}

func (conn PodManContext) StartService(containerName string) (err error) {
	if err := containers.Start(conn, containerName, nil); err != nil {
		return err
	}
	return nil
}

func (conn PodManContext) StopService(containerName string) (err error) {
	if err := containers.Stop(conn, containerName, nil); err != nil {
		return err
	}
	return nil
}

func (conn PodManContext) PauseService(containerName string) (err error) {
	return nil
}

func (conn PodManContext) UpdateService(containerName string) (err error) {
	return nil
}

func (conn PodManContext) ListNetworks() (networks []string, err error) {
	networkList, err := network.List(conn, nil)

	for _, network := range networkList {
		networks = append(networks, network.Name)
	}

	return
}

func (conn PodManContext) CreateNetwork(networkName string) (err error) {
	return nil
}

func (conn PodManContext) NetworkExists(networkName string) (exists bool, err error) {
	return network.Exists(conn, networkName, nil)
}
