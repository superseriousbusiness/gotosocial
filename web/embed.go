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

package web

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// HybridFS is a fs.FS that tries to read files from a specified location on disk,
// and falls back to assets embedded in the executable. Can and should be used for
// any assets that live in subdirectories of this module.
type HybridFS struct {
	embeddedBaseDir EmbeddedFileGroup
	hostBaseDir     string
}

// EmbeddedFileGroup represents a subdirectory of this module, i.e. web/template/
type EmbeddedFileGroup string

const (
	EmbeddedTemplates EmbeddedFileGroup = "template"
	EmbeddedAssets    EmbeddedFileGroup = "assets"
)

// NewHybridFS creates an fs.FS that combines reading from embedded files and from disk.
func NewHybridFS(group EmbeddedFileGroup, baseDir string) HybridFS {
	return HybridFS{
		embeddedBaseDir: group,
		hostBaseDir:     baseDir,
	}
}

func (h HybridFS) Open(name string) (f fs.File, err error) {
	// NOTE: we are responsible here for checking for tree traversal (as per
	// fs.FS interface definition).
	// Because we want to be able to use a hostBaseDir that is outside of PWD,
	// we don't verify ValidPath() on the full path (it would reject absolute
	// paths and relative paths containing `..`), but only the given asset
	// name. In other words: We trust the value of h.hostBaseDir to not contain
	// a malicious path.
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	// first try to read from host filesystem in the chosen hostBaseDir
	f, err = os.Open(filepath.Join(h.hostBaseDir, name))

	// fall back to buildtime embedded files
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		path := filepath.Join(string(h.embeddedBaseDir), name)
		f, err = embeddedFiles.Open(path)
	}

	return f, err
}
