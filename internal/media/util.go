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

const (
	// MIMEImage is the mime type for image
	MIMEImage = "image"
	// MIMEJpeg is the jpeg image mime type
	MIMEJpeg = "image/jpeg"
	// MIMEGif is the gif image mime type
	MIMEGif = "image/gif"
	// MIMEPng is the png image mime type
	MIMEPng = "image/png"

	// MIMEVideo is the mime type for video
	MIMEVideo = "video"
	// MIMEMp4 is the mp4 video mime type
	MIMEMp4 = "video/mp4"
	// MIMEMpeg is the mpeg video mime type
	MIMEMpeg = "video/mpeg"
	// MIMEWebm is the webm video mime type
	MIMEWebm = "video/webm"
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

// SupportedImageType checks mime type of an image against a slice of accepted types,
// and returns True if the mime type is accepted.
func SupportedImageType(mimeType string) bool {
	acceptedImageTypes := []string{
		MIMEJpeg,
		MIMEGif,
		MIMEPng,
	}
	for _, accepted := range acceptedImageTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// SupportedVideoType checks mime type of a video against a slice of accepted types,
// and returns True if the mime type is accepted.
func SupportedVideoType(mimeType string) bool {
	acceptedVideoTypes := []string{
		MIMEMp4,
		MIMEMpeg,
		MIMEWebm,
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
		MIMEGif,
		MIMEPng,
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
	case MIMEGif:
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

	return &imageAndMeta{
		image:  b,
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

func deriveImage(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case MIMEJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case MIMEPng:
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

	return &imageAndMeta{
		image:  b,
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
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
	case MIMEJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case MIMEPng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case MIMEGif:
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

	tiny := resize.Thumbnail(32, 32, thumb, resize.NearestNeighbor)
	bh, err := blurhash.Encode(4, 3, tiny)
	if err != nil {
		return nil, err
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, thumb, &jpeg.Options{
		Quality: 75,
	}); err != nil {
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

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(b []byte, contentType string) (*imageAndMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case MIMEPng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case MIMEGif:
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
func ParseMediaType(s string) (Type, error) {
	switch Type(s) {
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
func ParseMediaSize(s string) (Size, error) {
	switch Size(s) {
	case Small:
		return Small, nil
	case Original:
		return Original, nil
	case Static:
		return Static, nil
	}
	return "", fmt.Errorf("%s not a recognized MediaSize", s)
}
