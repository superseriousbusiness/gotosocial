package iptc

import (
	"os"
	"path"

	"github.com/dsoprea/go-logging"
)

var (
	testDataRelFilepath = "iptc.data"
)

var (
	moduleRootPath = ""
	assetsPath     = ""
)

func GetModuleRootPath() string {
	if moduleRootPath == "" {
		moduleRootPath = os.Getenv("IPTC_MODULE_ROOT_PATH")
		if moduleRootPath != "" {
			return moduleRootPath
		}

		currentWd, err := os.Getwd()
		log.PanicIf(err)

		currentPath := currentWd
		visited := make([]string, 0)

		for {
			tryStampFilepath := path.Join(currentPath, ".MODULE_ROOT")

			_, err := os.Stat(tryStampFilepath)
			if err != nil && os.IsNotExist(err) != true {
				log.Panic(err)
			} else if err == nil {
				break
			}

			visited = append(visited, tryStampFilepath)

			currentPath = path.Dir(currentPath)
			if currentPath == "/" {
				log.Panicf("could not find module-root: %v", visited)
			}
		}

		moduleRootPath = currentPath
	}

	return moduleRootPath
}

func GetTestAssetsPath() string {
	if assetsPath == "" {
		moduleRootPath := GetModuleRootPath()
		assetsPath = path.Join(moduleRootPath, "assets")
	}

	return assetsPath
}

func GetTestDataFilepath() string {
	assetsPath := GetTestAssetsPath()
	filepath := path.Join(assetsPath, testDataRelFilepath)

	return filepath
}
