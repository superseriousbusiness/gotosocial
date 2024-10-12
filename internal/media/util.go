// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-mimetypes"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// file represents one file
// with the given flag and perms.
type file struct {
	abs  string
	flag int
	perm os.FileMode
}

// allowFiles implements fs.FS to allow
// access to a specified slice of files.
type allowFiles []file

// Open implements fs.FS.
func (af allowFiles) Open(name string) (fs.File, error) {
	for _, file := range af {
		var (
			abs  = file.abs
			flag = file.flag
			perm = file.perm
		)

		// Allowed to open file
		// at absolute path.
		if name == file.abs {
			return os.OpenFile(abs, flag, perm)
		}

		// Check for other valid reads.
		thisDir, thisFile := path.Split(file.abs)

		// Allowed to read directory itself.
		if name == thisDir || name == "." {
			return os.OpenFile(thisDir, flag, perm)
		}

		// Allowed to read file
		// itself (at relative path).
		if name == thisFile {
			return os.OpenFile(abs, flag, perm)
		}
	}

	return nil, os.ErrPermission
}

// getExtension splits file extension from path.
func getExtension(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i+1:]
		}
	}
	return ""
}

// getMimeType returns a suitable mimetype for file extension.
func getMimeType(ext string) string {
	const defaultType = "application/octet-stream"
	return cmp.Or(mimetypes.MimeTypes[ext], defaultType)
}

// drainToTmp drains data from given reader into a new temp file
// and closes it, returning the path of the resulting temp file.
//
// Note that this function specifically makes attempts to unwrap the
// io.ReadCloser as much as it can to underlying type, to maximise
// chance that Linux's sendfile syscall can be utilised for optimal
// draining of data source to temporary file storage.
func drainToTmp(rc io.ReadCloser) (string, error) {
	defer rc.Close()

	// Open new temporary file.
	tmp, err := os.CreateTemp(
		os.TempDir(),
		"gotosocial-*",
	)
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	// Extract file path.
	path := tmp.Name()

	// Limited reader (if any).
	var lr *io.LimitedReader
	var limit int64

	// Reader type to use
	// for draining to tmp.
	rd := (io.Reader)(rc)

	// Check if reader is actually wrapped,
	// (as our http client wraps close func).
	rct, ok := rc.(*iotools.ReadCloserType)
	if ok {

		// Get unwrapped.
		rd = rct.Reader

		// Extract limited reader if wrapped.
		lr, limit = iotools.GetReaderLimit(rd)
	}

	// Drain reader into tmp.
	_, err = tmp.ReadFrom(rd)
	if err != nil {
		return path, err
	}

	// Check to see if limit was reached,
	// (produces more useful error messages).
	if lr != nil && lr.N <= 0 {
		err := fmt.Errorf("reached read limit %s", bytesize.Size(limit)) // #nosec G115 -- Just logging
		return path, gtserror.SetLimitReached(err)
	}

	return path, nil
}

// remove only removes paths if not-empty.
func remove(paths ...string) error {
	var errs []error
	for _, path := range paths {
		if path != "" {
			if err := os.Remove(path); err != nil {
				errs = append(errs, fmt.Errorf("error removing %s: %w", path, err))
			}
		}
	}
	return errors.Join(errs...)
}
