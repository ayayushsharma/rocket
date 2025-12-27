package containers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containers/podman/v6/pkg/bindings"
)

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
)


type connectionFileV2 struct {
    Connection struct {
        Default     string                          `json:"Default"`
        Connections map[string]connectionEntryV2    `json:"Connections"`
    } `json:"Connection"`
    // "Farm" may exist, we ignore it.
}

type connectionEntryV2 struct {
    URI      string `json:"URI"`
    Identity string `json:"Identity"`
    // Some builds add flags like IsMachine, TLS*, etc.
}

// --- Legacy schema (array of objects produced by `podman system connection list --format=json`) ---
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

    xdg := os.Getenv("XDG_CONFIG_HOME")
    var base string
    if xdg != "" {
        base = xdg
    } else {
        home, err := os.UserHomeDir()
        if err != nil {
            return "", fmt.Errorf("cannot resolve home dir: %w", err)
        }
        base = filepath.Join(home, ".config")
    }
    return filepath.Join(base, "containers", "podman-connections.json"), nil
}

// LoadDefaultConnection reads the file and returns (uri, identity).
func LoadDefaultConnection() (string, string, error) {
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

func DefaultSocketURI() (string, error) {
	uri, _, err := LoadDefaultConnection()
	return uri, err
}

func Connect() (context.Context, error) {
	socketURI, err := DefaultSocketURI()
	if err != nil {
		slog.Error("Couldn't connect to Podman", "error", err)
		slog.Error("Check if Podman service is running on the machine")
		return nil, err
	}

	conn, err := bindings.NewConnection(context.Background(), socketURI)

	if err != nil {
		slog.Error("Couldn't connect to podman service", "error", err)
		return nil, err
	}

	slog.Debug("Connected to podman")
	return conn, nil
}
