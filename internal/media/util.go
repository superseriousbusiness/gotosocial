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
	"image"
	"image/jpeg"
	"io"
	"os"

	"codeberg.org/gruf/go-mimetypes"
	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
)

// jpegDecode ...
func jpegDecode(filepath string) (image.Image, error) {
	// Open the file at given path.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	// Decode image from file.
	img, err := jpeg.Decode(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	// Close file now decoded into mem.
	if err := file.Close(); err != nil {
		return nil, err
	}

	return img, nil
}

// generateBlurhash ...
func generateBlurhash(filepath string) (string, error) {
	// Decode JPEG file at given path.
	img, err := jpegDecode(filepath)
	if err != nil {
		return "", err
	}

	// for generating blurhashes, it's more cost effective to
	// lose detail since it's blurry, so make a tiny version.
	tiny := imaging.Resize(img, 64, 64, imaging.NearestNeighbor)

	// Drop the larger image
	// ref as soon as possible
	// to allow GC to claim.
	img = nil //nolint

	// Generate blurhash for thumbnail.
	return blurhash.Encode(4, 3, tiny)
}

// getMimeType returns a suitable mimetype for file extension.
func getMimeType(ext string) string {
	const defaultType = "application/octet-stream"
	return cmp.Or(mimetypes.MimeTypes[ext], defaultType)
}

// drainToTmp drains data from given reader into a new temp file
// and closes it, returning the path of the resulting temp file.
func drainToTmp(r io.Reader) (string, error) {
	// Create new temp output.
	tmp, err := os.CreateTemp(
		os.TempDir(),
		"gotosocial-*",
	)
	if err != nil {
		return "", err
	}

	// Extract file path.
	path := tmp.Name()

	// Drain input reader into temporary file.
	if _, err = tmp.ReadFrom(r); err != nil {
		_ = tmp.Close()
		return path, err
	}

	// Close file now finished.
	return path, tmp.Close()
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
