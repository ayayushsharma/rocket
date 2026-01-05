package containers

type Container interface {
	PullImage(imageName string) error
	RemoveImage(imageName string) error

	CreateContainer(options ContainerConfig) error
	RemoveContainer(containerName string, force bool) error

	StartService(containerName string) error
	PauseService(containerName string) error
	StopService(containerName string) error
	UpdateService(containerName string) error

	ListNetworks() ([]string, error)
	CreateNetwork(networkName string) error
	NetworkExists(networkName string) (bool, error)
}
