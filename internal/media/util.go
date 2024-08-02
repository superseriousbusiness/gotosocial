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
	"io"
	"os"

	"golang.org/x/image/webp"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-mimetypes"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
)

// thumbSize returns the dimensions to use for an input
// image of given width / height, for its outgoing thumbnail.
// This attempts to maintains the original image aspect ratio.
func thumbSize(width, height int, aspect float32) (int, int) {
	const (
		maxThumbWidth  = 512
		maxThumbHeight = 512
	)

	switch {
	// Simplest case, within bounds!
	case width < maxThumbWidth &&
		height < maxThumbHeight:
		return width, height

	// Width is larger side.
	case width > height:
		// i.e. height = newWidth * (height / width)
		height = int(float32(maxThumbWidth) / aspect)
		return maxThumbWidth, height

	// Height is larger side.
	case height > width:
		// i.e. width = newHeight * (width / height)
		width = int(float32(maxThumbHeight) * aspect)
		return width, maxThumbHeight

	// Square.
	default:
		return maxThumbWidth, maxThumbHeight
	}
}

// webpDecode decodes the WebP at filepath into parsed image.Image.
func webpDecode(filepath string) (image.Image, error) {
	// Open the file at given path.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	// Decode image from file.
	img, err := webp.Decode(file)

	// Done with file.
	_ = file.Close()

	return img, err
}

// generateBlurhash generates a blurhash for JPEG at filepath.
func generateBlurhash(filepath string) (string, error) {
	// Decode JPEG file at given path.
	img, err := webpDecode(filepath)
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
	if lr != nil && !iotools.AtEOF(lr.R) {
		return path, fmt.Errorf("reached read limit %s", bytesize.Size(limit))
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
