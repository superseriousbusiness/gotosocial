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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
)

// media processing tmpdir.
var tmpdir = os.TempDir()

// file represents one file
// with the given flag and perms.
type file struct {
	abs  string // absolute file path, including root
	dir  string // containing directory of abs
	rel  string // relative to root, i.e. trim_prefix(abs, dir)
	flag int
	perm os.FileMode
}

// allowRead returns a new file{} for filepath permitted only to read.
func allowRead(filepath string) file {
	return newFile(filepath, os.O_RDONLY, 0)
}

// allowCreate returns a new file{} for filepath permitted to read / write / create.
func allowCreate(filepath string) file {
	return newFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}

// newFile returns a new instance of file{} for given path and open args.
func newFile(filepath string, flag int, perms os.FileMode) file {
	dir, rel := path.Split(filepath)
	return file{
		abs:  filepath,
		rel:  rel,
		dir:  dir,
		flag: flag,
		perm: perms,
	}
}

// allowFiles implements fs.FS to allow
// access to a specified slice of files.
type allowFiles []file

// Open implements fs.FS.
func (af allowFiles) Open(name string) (fs.File, error) {
	for _, file := range af {
		switch name {
		// Allowed to open file
		// at absolute path, or
		// relative as ffmpeg likes.
		case file.abs, file.rel:
			return os.OpenFile(file.abs, file.flag, file.perm)

		// Ffmpeg likes to read containing
		// dir as '.'. Allow RO access here.
		case ".":
			return openRead(file.dir)
		}
	}
	return nil, os.ErrPermission
}

// openRead opens the existing file at path for reads only.
func openRead(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, 0)
}

// openWrite opens the (new!) file at path for read / writes.
func openWrite(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
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

// drainToTmp drains data from given reader into a new temp file
// and closes it, returning the path of the resulting temp file.
//
// Note that this function specifically makes attempts to unwrap the
// io.ReadCloser as much as it can to underlying type, to maximise
// chance that Linux's sendfile syscall can be utilised for optimal
// draining of data source to temporary file storage.
func drainToTmp(rc io.ReadCloser) (string, error) {
	var tmp *os.File
	var err error

	// Close handles
	// on func return.
	defer func() {
		tmp.Close()
		rc.Close()
	}()

	// Open new temporary file.
	tmp, err = os.CreateTemp(
		tmpdir,
		"gotosocial-*",
	)
	if err != nil {
		return "", err
	}

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
