package containers

import "errors"

var ContainerAlreadyExistsErr = errors.New("container already exists")
var ContainerDoesntExistErr = errors.New("container does not exist")
