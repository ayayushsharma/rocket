package common


import (
	"strings"
	"fmt"
	"ayayushsharma/rocket/constants"
)

func CreateContainerName(
	applicationName,
	imageName string,
	imageVersion string,
) (containerName string) {
	imageName = strings.ReplaceAll(imageName, "/", "-")
	imageName = strings.ReplaceAll(imageName, "\\", "-")
	imageName = strings.ReplaceAll(imageName, ":", "-")
	imageName = strings.ReplaceAll(imageName, ".", "-")

	imageVersion = strings.ReplaceAll(imageName, ".", "-")

	containerName = fmt.Sprintf(
		"%s.%s.%s.%s",
		constants.ApplicationName,
		applicationName,
		imageName,
		imageVersion,
	)
	return
}

func ExtractImageName(imageUrl string) string {
	containsVersion := strings.Contains(imageUrl, ":")
	if containsVersion {
		versionIndex := strings.Index(imageUrl, ":")
		return imageUrl[0:versionIndex]
	}
	return imageUrl
}

func ExtractImageVersion(imageUrl string) string {
	containsVersion := strings.Contains(imageUrl, ":")
	if containsVersion {
		versionIndex := strings.Index(imageUrl, ":")
		return imageUrl[versionIndex + 1:]
	}
	return "latest"
}
