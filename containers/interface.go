package containers

type ContainerManager interface {
	ImageExists(imageName string) (bool, error)
	PullImage(imageName string) error
	RemoveImage(imageName string) error

	CreateContainer(options Config) error
	RemoveContainer(containerName string, force bool) error

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
