/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package embedded

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/web"
)

var ListEmbeddedFiles action.GTSAction = func(ctx context.Context) error {
	return fs.WalkDir(web.EmbeddedFiles, ".", func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			fmt.Println(path)
		}
		return nil
	})
}

func ViewEmbeddedFile(sourcePath string) error {
	f, err := web.EmbeddedFiles.Open(sourcePath)
	if err != nil {
		return err
	}
	content, err := io.ReadAll(f)
	fmt.Printf("%s", content)
	return err
}

func ExtractEmbeddedFile(sourcePath, targetBaseDir string) error {
	if targetBaseDir == "" {
		// select target dir based on configuration
		if strings.HasPrefix(sourcePath, string(web.EmbeddedAssets)) {
			targetBaseDir = config.GetWebAssetBaseDir()
		} else if strings.HasPrefix(sourcePath, string(web.EmbeddedTemplates)) {
			targetBaseDir = config.GetWebTemplateBaseDir()
		}
		// remove the duplicate template / assets path element (it's part of `sourcePath` already)
		targetBaseDir = filepath.Dir(strings.TrimSuffix(targetBaseDir, "/"))
	}

	targetPath := filepath.Join(targetBaseDir, sourcePath)
	_, err := os.Stat(targetPath)
	if err == nil { //nolint:gocritic,ifElseChain
		// target file seems to exist, move existing content to *.bak first
		if err := createBackup(targetPath); err != nil {
			return fmt.Errorf("could not create backup of '%s': %s", targetPath, err)
		}
	} else if errors.Is(err.(*fs.PathError).Err, fs.ErrNotExist) { //nolint:forcetypeassert
		// ErrNotExist is not an error, but everything else is.
		// We still need to ensure the parent path exists.
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("could not create parent directory for '%s': %s", targetPath, err)
		}
	} else {
		return fmt.Errorf("could not check for existing file '%s': %s", targetPath, err)
	}

	// open files & copy content
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("could not open new file '%s': %s", targetPath, err)
	}
	defer targetFile.Close()
	sourceFile, err := web.EmbeddedFiles.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open embedded file '%s': %s", sourcePath, err)
	}
	defer sourceFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return fmt.Errorf("could not copy to '%s': %s", targetFile.Name(), err)
	}
	return nil
}

func createBackup(path string) error {
	original, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("could not open existing file '%s': %s", path, err)
	}
	defer original.Close()
	backup, err := os.Create(path + ".bak")
	if err != nil {
		return fmt.Errorf("could not open new backup file '%s': %s", path, err)
	}
	defer backup.Close()
	_, err = io.Copy(backup, original)
	return err
}
