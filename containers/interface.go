package containers

type ContainerManager interface {
	PullImage(imageName string) error
	RemoveImage(imageName string) error
	ImageExists(imageName string) (bool, error)

	// ListContainers() ([]string, error)
	CreateContainer(options Config) error
	RemoveContainer(containerName string, force bool) error
	ContainerExists(containerName string) (bool, error)

	StartService(containerName string) error
	PauseService(containerName string) error
	StopService(containerName string) error
	UpdateService(containerName string) error

	ListNetworks() ([]string, error)
	CreateNetwork(networkName string) error
	NetworkExists(networkName string) (bool, error)
}

func Manager() (manager ContainerManager, err error) {
	return connectPodman()
}
