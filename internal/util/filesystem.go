/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package util

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// If the file already exists, this function will perform a write and a read on that file,
// then clean up after itself, reporting any errors along the way.
// If the file does not already exist, it will instead try to create and then read from a separate test file
// in the same parent folder, then clean up after itself, reporting any errors along the way.
func FilePathOk(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("error obtaining absolute path from '%s': %s", path, err)
	}

	// these file permissions are not used, as the O_CREATE flag is not passed,
	// however they are required by the os.OpenFile API
	onlyOwnerCanReadAndWrite := fs.FileMode(0600)

	theFile, err := os.OpenFile(absPath, os.O_RDWR, onlyOwnerCanReadAndWrite)
	if err != nil {

		if os.IsNotExist(err) {

			// The file does not exist yet.
			// So check the permissions of the parent folder instead.
			folderPath := filepath.Dir(absPath)
			ioTestFilename := fmt.Sprintf("%s.iotest", absPath)
			err = ioutil.WriteFile(ioTestFilename, []byte("check_to_see_if_file_is_writable"), onlyOwnerCanReadAndWrite)
			if err != nil {
				return fmt.Errorf("I cannot write a file inside '%s': %s", folderPath, err)
			}
			_, err = ioutil.ReadFile(ioTestFilename)
			if err != nil {
				return fmt.Errorf("I cannot read a file inside '%s': %s", folderPath, err)
			}

			err = os.Remove(ioTestFilename)
			if err != nil {
				return fmt.Errorf("I can read and write a file inside '%s', but deleting it failed: %s", folderPath, err)
			}

			return nil
		}

		return err
	}

	theFile.Close()

	return nil
}
