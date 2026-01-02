package containers

type Container interface {
	PullImage(imageName string) error
	RemoveImage(imageName string) error

	CreateContainer(
		options ContainerCreateOptions,
	) error

	StartService(containerName string) error
	PauseService(containerName string) error
	StopService(containerName string) error
	UpdateService(containerName string) error

	ListNetworks() ([]string, error)
	CreateNetwork(networkName string) error
	NetworkExists(networkName string) (bool, error)
}
