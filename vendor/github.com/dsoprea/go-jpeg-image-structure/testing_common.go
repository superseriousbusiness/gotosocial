package jpegstructure

import (
	"os"
	"path"

	"github.com/dsoprea/go-logging"
)

var (
	testImageRelFilepath = "NDM_8901.jpg"
)

var (
	moduleRootPath = ""
	assetsPath     = ""
)

// GetModuleRootPath returns the root-path of the module.
func GetModuleRootPath() string {
	if moduleRootPath == "" {
		moduleRootPath = os.Getenv("JPEG_MODULE_ROOT_PATH")
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

// GetTestAssetsPath returns the path of the test-assets.
func GetTestAssetsPath() string {
	if assetsPath == "" {
		moduleRootPath := GetModuleRootPath()
		assetsPath = path.Join(moduleRootPath, "assets")
	}

	return assetsPath
}

// GetTestImageFilepath returns the file-path of the common test-image.
func GetTestImageFilepath() string {
	assetsPath := GetTestAssetsPath()
	filepath := path.Join(assetsPath, testImageRelFilepath)

	return filepath
}
