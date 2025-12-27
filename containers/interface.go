package containers


type Container interface {
	PullImage(imageName string)
	StartService(
		imageName string,
		containerName string,
		endpointPrefix string,
		applicationName string,
	)
	PauseService(containerName string)
	StopService(containerName string)
}
