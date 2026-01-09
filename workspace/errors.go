package workspace

import (
	"errors"
	"fmt"
)

type AppAlreadyRegisteredErr struct {
	ContainerName string
}

func (e *AppAlreadyRegisteredErr) Error() string {
	return fmt.Sprintf("App already registered as: %s", e.ContainerName)
}

var AppNotRegisteredErr error = errors.New("This app is not registered")
var NoAppSelectedErr error = errors.New("No app selected for registration")
