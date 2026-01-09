package common

import (
	"fmt"
	"strings"

	"ayayushsharma/rocket/constants"
)

// Creates application name to be used for registering the application
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

// Extract image name from artifactory URl
func ExtractImageName(imageUrl string) string {
	containsVersion := strings.Contains(imageUrl, ":")
	if containsVersion {
		versionIndex := strings.Index(imageUrl, ":")
		return imageUrl[0:versionIndex]
	}
	return imageUrl
}

// Extracts image version for am image URL.
// Defaults to "latest" if version is not supplied
func ExtractImageVersion(imageUrl string) string {
	containsVersion := strings.Contains(imageUrl, ":")
	if containsVersion {
		versionIndex := strings.Index(imageUrl, ":")
		return imageUrl[versionIndex+1:]
	}
	return "latest"
}

// Concatenates Image with it's version
func ImageWithVersion(imageUrl string, imageVersion string) string {
	imageVersion = strings.TrimSpace(imageVersion)
	if imageVersion == "" {
		return imageUrl
	}
	return fmt.Sprintf("%s:%s", imageUrl, imageVersion)
}

// Returns shortened user name of the application after removing "rocket-"
// from the beginning
func ShortenAppName(appName string) string {
	if len(appName) <= 7 {
		return appName
	}
	if appName[0:7] == "rocket-" {
		return strings.ReplaceAll(appName, "rocket-", "")
	}
	return appName
}

// Returns the complete name of the application that can be used internally.
// All rocket apps are prefixes with "rocket-"
func CompleteAppName(appName string) string {
	if len(appName) <= 7 {
		return "rocket-" + appName
	}
	if appName[0:7] == "rocket-" {
		return appName
	}
	return "rocket-" + appName
}
