package resources

import (
	"embed"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"ayayushsharma/rocket/constants"
)

//go:embed home-page/* nginx/*
var staticFiles embed.FS

type resourceFile struct {
	RelativePath string
	Data         []byte
}

var files []string

func ResourceFiles() ([]resourceFile, error) {
	resourceFiles := []resourceFile{}

	for _, fileName := range files {
		content, err := staticFiles.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		resourceFiles = append(resourceFiles, resourceFile{
			RelativePath: fileName,
			Data:         content,
		})
	}

	return resourceFiles, nil
}

func CheckAll() (ok bool) {
	for _, filePath := range files {
		absFilePath := filepath.Join(constants.AppStateDir, filePath)
		_, err := os.Stat(absFilePath)
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}

func SyncAll() (err error) {
	// remove files for new files to take place
	for _, filePath := range files {
		_ = os.Remove(filePath)
	}

	resourceFiles, err := ResourceFiles()
	if err != nil {
		return err
	}

	for _, resrcFile := range resourceFiles {
		filePath := filepath.Join(constants.AppStateDir, resrcFile.RelativePath)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		err = os.WriteFile(filePath, resrcFile.Data, 0644)
		if err != nil {
			return
		}
	}

	slog.Debug("Synced the app data")
	return nil
}

func init() {
	files = []string{
		"nginx/nginx.conf",
		"home-page/index.html",
		"home-page/custom_50x.html",
		"home-page/static/script.js",
		"home-page/static/style.css",
	}
}
