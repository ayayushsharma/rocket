package containers

// for loading config
type Config struct {
	// name of the application that will be displayed to the user on the GUI
	ApplicationName string

	// Name of the container. If not supplied, custom default naming scheme will
	// be used
	ContainerName string

	// name of the imageName to be pulled from artifactory. Entire URL can be
	// used here except the version
	ImageURL string

	// Version of the image to be pulled from artifactory
	ImageVersion string

	// subdomain that the container will use to direct network to the container
	SubDomain string

	// name of the `network` space in the container runners to isolate
	// the application
	NetworkName string

	// mounts host directories to containers
	// mountDirs["HOST_DIR"] = "CONTAINER_DIR"
	MountDirs map[string]string

	// bind host ports to container ports
	// bindPorts["HOST_PORT"] = "CONTAINER_PORT"
	BindPorts map[int]int

	// environment variables and their values to be passed to containers
	EnvValues map[string]string

	// env vars to be passed from host machine directly to pods
	EnvVars []string

	// http port to export
	ExposeHttpPort int
}
