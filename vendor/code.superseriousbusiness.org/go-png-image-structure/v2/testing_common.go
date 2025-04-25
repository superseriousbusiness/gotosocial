package pngstructure

import (
	"fmt"
	"os"
	"path"
)

var (
	assetsPath = "assets"
)

func getModuleRootPath() (string, error) {
	moduleRootPath := os.Getenv("PNG_MODULE_ROOT_PATH")
	if moduleRootPath != "" {
		return moduleRootPath, nil
	}

	currentWd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	currentPath := currentWd
	visited := make([]string, 0)

	for {
		tryStampFilepath := path.Join(currentPath, ".MODULE_ROOT")

		_, err := os.Stat(tryStampFilepath)
		if err != nil && !os.IsNotExist(err) {
			return "", err
		} else if err == nil {
			break
		}

		visited = append(visited, tryStampFilepath)

		currentPath = path.Dir(currentPath)
		if currentPath == "/" {
			return "", fmt.Errorf("could not find module-root: %v", visited)
		}
	}

	return currentPath, nil
}

func getTestAssetsPath() (string, error) {
	if assetsPath == "" {
		moduleRootPath, err := getModuleRootPath()
		if err != nil {
			return "", err
		}

		assetsPath = path.Join(moduleRootPath, "assets")
	}

	return assetsPath, nil
}

func getTestBasicImageFilepath() (string, error) {
	assetsPath, err := getTestAssetsPath()
	if err != nil {
		return "", err
	}

	return path.Join(assetsPath, "libpng.png"), nil
}

func getTestExifImageFilepath() (string, error) {
	assetsPath, err := getTestAssetsPath()
	if err != nil {
		return "", err
	}

	return path.Join(assetsPath, "exif.png"), nil
}
