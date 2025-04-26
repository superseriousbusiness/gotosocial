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
	"context"
	"encoding/binary"
	"image/jpeg"
	"io"
	"os"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-byteutil"
)

const (
	// image magic header bytes.
	magicJPEG = "\xff\xd8\xff"
)

// probe will first attempt to probe the file at path using native Go code
// (for performance), but falls back to using ffprobe to retrieve media details.
func probe(ctx context.Context, filepath string) (*result, error) {
	// Open input file at given path.
	file, err := os.Open(filepath)
	if err != nil {
		return nil, gtserror.Newf("error opening file %s: %w", filepath, err)
	}

	// Close on return.
	defer file.Close()

	// Byte buf to check for
	// file header magic bytes.
	buf := make([]byte, 3)

	// Read file header into buffer.
	_, err = io.ReadFull(file, buf)
	if err != nil {
		return nil, gtserror.Newf("error reading file %s: %w", filepath, err)
	}

	switch {
	// Attempt to probe JPEG types
	// separately, to save calls into
	// WebAssembly for a common image.
	case string(buf[:len(magicJPEG)]) == magicJPEG:
		log.Debug(ctx, "probing jpeg")
		return probeJPEG(file)

	default:
		// Close BEFORE
		// pass to ffprobe.
		_ = file.Close()

		// For everything else, fall back
		// to calling ffprobe on input file.
		log.Debug(ctx, "ffprobing file")
		return ffprobe(ctx, filepath)
	}
}

// probeJPEG decodes the given file as JPEG and determines
// image details from the decoded JPEG using native Go code.
func probeJPEG(file *os.File) (*result, error) {
	// Attempt to decode JPEG, adding back hdr magic.
	cfg, err := jpeg.DecodeConfig(io.MultiReader(
		strings.NewReader(magicJPEG),
		file,
	))
	if err != nil {
		return nil, gtserror.Newf("error decoding file %s: %w", file.Name(), err)
	}

	// Jump back to file start.
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, gtserror.Newf("error seeking in file %s: %w", file.Name(), err)
	}

	// Read orientation data from EXIF.
	orientation := readOrientation(file)

	// Setup result as if
	// ffprobe'd resulting in
	// JPEG file container.
	var res result
	res.format = "image2"

	// Set image orientation data.
	res.orientation = orientation

	// Extract image details.
	res.video = []videoStream{{
		stream: stream{codec: "mjpeg"},
		width:  cfg.Width,
		height: cfg.Height,

		// setting a pixel color format
		// doesn't matter for JPEG, as we
		// don't bother even using it.
		pixfmt: "",
	}}

	return &res, nil
}

// readOrientation reads orientation EXIF
// data (if it even exists) from image file.
//
// copied from github.com/disintegration/imaging
// but modified to optimize discard operations.
func readOrientation(r *os.File) int {
	const (
		markerAPP1     = 0xffe1
		exifHeader     = 0x45786966
		byteOrderBE    = 0x4d4d
		byteOrderLE    = 0x4949
		orientationTag = 0x0112
	)

	// Setup a discard read buffer.
	buf := new(byteutil.Buffer)
	buf.Guarantee(32)

	// discard simply reads into buf.
	discard := func(n int) error {
		buf.Guarantee(n) // ensure big enough
		_, err := io.ReadFull(r, buf.B[:n])
		return err
	}

	// Skip past JPEG SOI marker.
	if err := discard(2); err != nil {
		return orientationUnspecified
	}

	// Find JPEG
	// APP1 marker.
	for {
		var marker, size uint16

		if err := binary.Read(r, binary.BigEndian, &marker); err != nil {
			return orientationUnspecified
		}

		if err := binary.Read(r, binary.BigEndian, &size); err != nil {
			return orientationUnspecified
		}

		if marker>>8 != 0xff {
			return orientationUnspecified // Invalid JPEG marker.
		}

		if marker == markerAPP1 {
			break
		}

		if size < 2 {
			return orientationUnspecified // Invalid block size.
		}

		if err := discard(int(size - 2)); err != nil {
			return orientationUnspecified
		}
	}

	// Check if EXIF
	// header is present.
	var header uint32

	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return orientationUnspecified
	}

	if header != exifHeader {
		return orientationUnspecified
	}

	if err := discard(2); err != nil {
		return orientationUnspecified
	}

	// Read byte
	// order info.
	var (
		byteOrderTag uint16
		byteOrder    binary.ByteOrder
	)

	if err := binary.Read(r, binary.BigEndian, &byteOrderTag); err != nil {
		return orientationUnspecified
	}

	switch byteOrderTag {
	case byteOrderBE:
		byteOrder = binary.BigEndian
	case byteOrderLE:
		byteOrder = binary.LittleEndian
	default:
		return orientationUnspecified // Invalid byte order flag.
	}

	if err := discard(2); err != nil {
		return orientationUnspecified
	}

	// Skip the
	// EXIF offset.
	var offset uint32

	if err := binary.Read(r, byteOrder, &offset); err != nil {
		return orientationUnspecified
	}

	if offset < 8 {
		return orientationUnspecified // Invalid offset value.
	}

	if err := discard(int(offset - 8)); err != nil {
		return orientationUnspecified
	}

	// Read the
	// number of tags.
	var numTags uint16

	if err := binary.Read(r, byteOrder, &numTags); err != nil {
		return orientationUnspecified
	}

	// Find the orientation tag.
	for i := 0; i < int(numTags); i++ {
		var tag uint16

		if err := binary.Read(r, byteOrder, &tag); err != nil {
			return orientationUnspecified
		}

		if tag != orientationTag {
			if err := discard(10); err != nil {
				return orientationUnspecified
			}
			continue
		}

		if err := discard(6); err != nil {
			return orientationUnspecified
		}

		var val uint16

		if err := binary.Read(r, byteOrder, &val); err != nil {
			return orientationUnspecified
		}

		if val < 1 || val > 8 {
			return orientationUnspecified // Invalid tag value.
		}

		return int(val)
	}

	// Missing orientation tag.
	return orientationUnspecified
}
