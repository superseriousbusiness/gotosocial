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

// supportedVideoType checks mime type of a video against a slice of accepted types,
// and returns True if the mime type is accepted.
func supportedVideoType(mimeType string) bool {
	acceptedVideoTypes := []string{
		"video/mp4",
		"video/mpeg",
		"video/webm",
	}
	for _, accepted := range acceptedVideoTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// supportedEmojiType checks that the content type is image/png -- the only type supported for emoji.
func supportedEmojiType(mimeType string) bool {
	acceptedEmojiTypes := []string{
		"image/gif",
		"image/png",
	}
	for _, accepted := range acceptedEmojiTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// purgeExif is a little wrapper for the action of removing exif data from an image.
// Only pass pngs or jpegs to this function.
func purgeExif(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, errors.New("passed image was not valid")
	}

	clean, err := exifremove.Remove(b)
	if err != nil {
		return nil, fmt.Errorf("could not purge exif from image: %s", err)
	}
	if len(clean) == 0 {
		return nil, errors.New("purged image was not valid")
	}
	return clean, nil
}

func deriveGif(b []byte, extension string) (*imageAndMeta, error) {
	var g *gif.GIF
	var err error
	switch extension {
	case "image/gif":
		g, err = gif.DecodeAll(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("extension %s not recognised", extension)
	}

	// use the first frame to get the static characteristics
	width := g.Config.Width
	height := g.Config.Height
	size := width * height
	aspect := float64(width) / float64(height)

	bh, err := blurhash.Encode(4, 3, g.Image[0])
	if err != nil || bh == "" {
		return nil, err
	}

	out := &bytes.Buffer{}
	if err := gif.EncodeAll(out, g); err != nil {
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

func deriveImage(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
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
	default:
		return nil, fmt.Errorf("content type %s not recognised", contentType)
	}

	width := i.Bounds().Size().X
	height := i.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	bh, err := blurhash.Encode(4, 3, i)
	if err != nil {
		return nil, err
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

// deriveThumbnail returns a byte slice and metadata for a thumbnail of width x and height y,
// of a given jpeg, png, or gif, or an error if something goes wrong.
//
// Note that the aspect ratio of the image will be retained,
// so it will not necessarily be a square, even if x and y are set as the same value.
func deriveThumbnail(b []byte, contentType string, x uint, y uint) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
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
		return nil, fmt.Errorf("content type %s not recognised", contentType)
	}

	thumb := resize.Thumbnail(x, y, i, resize.NearestNeighbor)
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

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
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
		return nil, fmt.Errorf("content type %s not allowed for emoji", contentType)
	}

	out := &bytes.Buffer{}
	if err := png.Encode(out, i); err != nil {
		return nil, err
	}
	return &imageAndMeta{
		image: out.Bytes(),
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

// ParseMediaType converts s to a recognized MediaType, or returns an error if unrecognized
func ParseMediaType(s string) (MediaType, error) {
	switch MediaType(s) {
	case Attachment:
		return Attachment, nil
	case Header:
		return Header, nil
	case Avatar:
		return Avatar, nil
	case Emoji:
		return Emoji, nil
	}
	return "", fmt.Errorf("%s not a recognized MediaType", s)
}

// ParseMediaSize converts s to a recognized MediaSize, or returns an error if unrecognized
func ParseMediaSize(s string) (MediaSize, error) {
	switch MediaSize(s) {
	case Small:
		return Small, nil
	case Original:
		return Original, nil
	case Static:
		return Static, nil
	}
	return "", fmt.Errorf("%s not a recognized MediaSize", s)
}
