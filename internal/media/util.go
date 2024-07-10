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

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-iotools"
	"codeberg.org/gruf/go-mimetypes"
	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// thumbSize returns the dimensions to use for an input
// image of given width / height, for its outgoing thumbnail.
// This maintains the original image aspect ratio.
func thumbSize(width, height int) (int, int) {
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
		p := float32(width) / float32(maxThumbWidth)
		return maxThumbWidth, int(float32(height) / p)

	// Height is larger side.
	case height > width:
		p := float32(height) / float32(maxThumbHeight)
		return int(float32(width) / p), maxThumbHeight

	// Square.
	default:
		return maxThumbWidth, maxThumbHeight
	}
}

// jpegDecode decodes the JPEG at filepath into parsed image.Image.
func jpegDecode(filepath string) (image.Image, error) {
	// Open the file at given path.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	// Decode image from file.
	img, err := jpeg.Decode(file)

	// Done with file.
	_ = file.Close()

	return img, err
}

// generateBlurhash generates a blurhash for JPEG at filepath.
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
//
// Note that this function specifically makes attempts to unwrap the
// io.ReadCloser as much as it can to underlying type, to maximise
// chance that Linux's sendfile syscall can be utilised for optimal
// draining of data source to temporary file storage.
func drainToTmp(rc io.ReadCloser) (string, error) {
	tmp, err := os.CreateTemp(os.TempDir(), "gotosocial-*")
	if err != nil {
		return "", err
	}

	// Close readers
	// on func return.
	defer tmp.Close()
	defer rc.Close()

	// Extract file path.
	path := tmp.Name()

	// Check for a
	// reader limit.
	var limit int64
	limit = -1

	// Reader type to use
	// for draining to tmp.
	rd := (io.Reader)(rc)

	// Check if reader is actually wrapped,
	// (as our http client wraps close func).
	rct, ok := rc.(*iotools.ReadCloserType)
	if ok {
		rd = rct.Reader

		// Extract limit if set on reader.
		_, limit = iotools.GetReaderLimit(rd)
	}

	// Drain reader into tmp.
	n, err := tmp.ReadFrom(rd)
	if err != nil {
		return path, err
	}

	// Check to see if limit was reached.
	if n == limit && !iotools.AtEOF(rd) {
		return path, gtserror.Newf("reached read limit %s", bytesize.Size(limit))
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
