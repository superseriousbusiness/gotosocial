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

package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/buckket/go-blurhash"
	"github.com/h2non/filetype"
	"github.com/nfnt/resize"
	"github.com/superseriousbusiness/exifremove/pkg/exifremove"
)

// parseContentType parses the MIME content type from a file, returning it as a string in the form (eg., "image/jpeg").
// Returns an error if the content type is not something we can process.
func parseContentType(content []byte) (string, error) {
	head := make([]byte, 261)
	_, err := bytes.NewReader(content).Read(head)
	if err != nil {
		return "", fmt.Errorf("could not read first magic bytes of file: %s", err)
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return "", err
	}

	if kind == filetype.Unknown {
		return "", errors.New("filetype unknown")
	}

	return kind.MIME.Value, nil
}

// supportedImageType checks mime type of an image against a slice of accepted types,
// and returns True if the mime type is accepted.
func supportedImageType(mimeType string) bool {
	acceptedImageTypes := []string{
		"image/jpeg",
		"image/gif",
		"image/png",
	}
	for _, accepted := range acceptedImageTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// purgeExif is a little wrapper for the action of removing exif data from an image.
// Only pass pngs or jpegs to this function.
func purgeExif(b []byte) ([]byte, error) {
	if b == nil || len(b) == 0 {
		return nil, errors.New("passed image was not valid")
	}

	clean, err := exifremove.Remove(b)
	if err != nil {
		return nil, fmt.Errorf("could not purge exif from image: %s", err)
	}
	if clean == nil || len(clean) == 0 {
		return nil, errors.New("purged image was not valid")
	}
	return clean, nil
}

func deriveImage(b []byte, extension string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch extension {
	case "image/jpeg":
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case "image/png":
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case "image/gif":
		i, err = gif.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("extension %s not recognised", extension)
	}

	width := i.Bounds().Size().X
	height := i.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)
	bh, err := blurhash.Encode(4, 3, i)
	if err != nil {
		return nil, fmt.Errorf("error generating blurhash: %s", err)
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, i, nil); err != nil {
		return nil, err
	}
	return &imageAndMeta{
		image:    out.Bytes(),
		width:    width,
		height:   height,
		size:     size,
		aspect:   aspect,
		blurhash: bh,
	}, nil
}

// deriveThumbnailFromImage returns a byte slice and metadata for a 256-pixel-width thumbnail
// of a given jpeg, png, or gif, or an error if something goes wrong.
//
// Note that the aspect ratio of the image will be retained,
// so it will not necessarily be a square.
func deriveThumbnail(b []byte, extension string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch extension {
	case "image/jpeg":
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case "image/png":
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case "image/gif":
		i, err = gif.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("extension %s not recognised", extension)
	}

	thumb := resize.Thumbnail(256, 256, i, resize.NearestNeighbor)
	width := thumb.Bounds().Size().X
	height := thumb.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, thumb, nil); err != nil {
		return nil, err
	}
	return &imageAndMeta{
		image:  out.Bytes(),
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

type imageAndMeta struct {
	image    []byte
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string
}
