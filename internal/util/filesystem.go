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
	"os"
	"path/filepath"
	"syscall"
)

func FilePathExistsAndIsReadWritable(path string) (exists bool, err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	parentFolderInfo, err := os.Stat(filepath.Dir(absPath))
	if err != nil {
		return false, fmt.Errorf("could not determine parent folder permissions for '%s': %s", absPath, err)
	}
	parentFolderStat := parentFolderInfo.Sys().(*syscall.Stat_t)
	processUserOwnsFolder := os.Geteuid() == int(parentFolderStat.Uid)
	processIsRunningAsRoot := os.Geteuid() == 0
	processHasGroupMatchForFolder := os.Getegid() == int(parentFolderStat.Gid)
	folderAllowsOwnerToReadAndWrite := (parentFolderInfo.Mode() & 0600) == 0600
	folderAllowsGroupMembersToReadAndWrite := (parentFolderInfo.Mode() & 0060) == 0060
	folderAllowsAnyoneToReadAndWrite := (parentFolderInfo.Mode() & 0006) == 0006
	readWritableBecauseUser := processUserOwnsFolder && folderAllowsOwnerToReadAndWrite
	// could use os.Getgroups() here to determine if the process user has membership in the group for the folder...
	// but does it really matter that much just for a warning message? 99% of the time the process UID should own the folder anyways
	readWritableBecauseGroup := processHasGroupMatchForFolder && folderAllowsGroupMembersToReadAndWrite

	if !(processIsRunningAsRoot || readWritableBecauseUser || readWritableBecauseGroup || folderAllowsAnyoneToReadAndWrite) {
		extraInfo := ""
		if !processUserOwnsFolder {
			extraInfo = fmt.Sprintf(": GoToSocial is running as user id %d, however the folder is owned by user %d", os.Geteuid(), int(parentFolderStat.Uid))
		} else if !folderAllowsOwnerToReadAndWrite {
			extraInfo = fmt.Sprintf(": GoToSocial (user id %d) owns the folder, however the owner is not allowed to read and write", os.Geteuid())
		}

		return true, fmt.Errorf("parent folder permissions for '%s' do not seem to allow us to read and write%s", absPath, extraInfo)
	}

	return true, nil
}
