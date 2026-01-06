package common

import (
	"ayayushsharma/rocket/constants"
	"fmt"
	"strings"
)

func CreateContainerName(
	applicationName,
	imageName string,
	imageVersion string,
) (containerName string) {
	imageNameSplit := strings.Split(imageName, "/")
	imageName = imageNameSplit[len(imageNameSplit)-1]
	imageName = strings.ReplaceAll(imageName, "\\", "_")
	imageName = strings.ReplaceAll(imageName, ":", "_")
	imageName = strings.ReplaceAll(imageName, ".", "_")

	imageVersion = strings.ReplaceAll(imageVersion, ".", "_")

	containerName = fmt.Sprintf(
		"%s-%s-%s",
		constants.ApplicationName,
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
		return imageUrl[versionIndex+1:]
	}
	return "latest"
}
