package containers

// for loading config
type ContainerConfigV1 struct {
	// name of the application that will be displayed to the user on the GUI
	applicationName string

	// Name of the container. If not supplied, custom default naming scheme will
	// be used
	containerName string

	// name of the imageName to be pulled from artifactory. Entire URL can be
	// used here except the version
	imageName string

	// Version of the image to be pulled from artifactory
	imageVersion string

	// subdomain that the container will use to direct network to the container
	subdomain string

	// name of the `network` space in the container runners to isolate
	// the application
	networkName string

	// mounts host directories to containers
	// mountDirs["HOST_DIR"] = "CONTAINER_DIR"
	mountDirs map[string]string

	// bind host ports to container ports
	// bindPorts["HOST_PORT"] = "CONTAINER_PORT"
	bindPorts map[int]int
}

// for configuring containers
type ContainerCreateOptions struct {
	ImageName       string
	ContainerName   string
	ApplicationName string
	SubDomain       string
	NetworkName     string
	MountDirs       map[string]string
	BindPorts       map[int]int
	EnvValues       map[string]string
	EnvVars         []string
}
