package workspace

import (
	"ayayushsharma/rocket/containers"
)

type workspaceSchema struct {
	Applications map[string]containers.Config `json:"applications"`
}

type routerData struct {
	ContainerURL string
	AppName      string
	Description  string
}
