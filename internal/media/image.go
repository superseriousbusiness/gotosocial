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

package media

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/buckket/go-blurhash"
	"github.com/nfnt/resize"
)

const (
	thumbnailMaxWidth  = 512
	thumbnailMaxHeight = 512
)

type imageMeta struct {
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string // defined only for calls to deriveThumbnail if createBlurhash is true
	small    []byte // defined only for calls to deriveStaticEmoji or deriveThumbnail
}

func decodeGif(r io.Reader) (*imageMeta, error) {
	gif, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}

	// use the first frame to get the static characteristics
	width := gif.Config.Width
	height := gif.Config.Height
	size := width * height
	aspect := float64(width) / float64(height)

	return &imageMeta{
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

func decodeImage(r io.Reader, contentType string) (*imageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(r)
	case mimeImagePng:
		i, err = StrippedPngDecode(r)
	default:
		err = fmt.Errorf("content type %s not recognised", contentType)
	}

	if err != nil {
		return nil, err
	}

	if i == nil {
		return nil, errors.New("processed image was nil")
	}

	width := i.Bounds().Size().X
	height := i.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	return &imageMeta{
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

// deriveThumbnail returns a byte slice and metadata for a thumbnail
// of a given jpeg, png, or gif, or an error if something goes wrong.
//
// If createBlurhash is true, then a blurhash will also be generated from a tiny
// version of the image. This costs precious CPU cycles, so only use it if you
// really need a blurhash and don't have one already.
//
// If createBlurhash is false, then the blurhash field on the returned ImageAndMeta
// will be an empty string.
func deriveThumbnail(r io.Reader, contentType string, createBlurhash bool) (*imageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(r)
	case mimeImagePng:
		i, err = StrippedPngDecode(r)
	case mimeImageGif:
		i, err = gif.Decode(r)
	default:
		err = fmt.Errorf("content type %s can't be thumbnailed", contentType)
	}

	if err != nil {
		return nil, fmt.Errorf("error decoding image as %s: %s", contentType, err)
	}

	if i == nil {
		return nil, errors.New("processed image was nil")
	}

	thumb := resize.Thumbnail(thumbnailMaxWidth, thumbnailMaxHeight, i, resize.NearestNeighbor)
	width := thumb.Bounds().Size().X
	height := thumb.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	im := &imageMeta{
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}

	if createBlurhash {
		// for generating blurhashes, it's more cost effective to lose detail rather than
		// pass a big image into the blurhash algorithm, so make a teeny tiny version
		tiny := resize.Thumbnail(32, 32, thumb, resize.NearestNeighbor)
		bh, err := blurhash.Encode(4, 3, tiny)
		if err != nil {
			return nil, fmt.Errorf("error creating blurhash: %s", err)
		}
		im.blurhash = bh
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, thumb, &jpeg.Options{
		// Quality isn't extremely important for thumbnails, so 75 is "good enough"
		Quality: 75,
	}); err != nil {
		return nil, fmt.Errorf("error encoding thumbnail: %s", err)
	}
	im.small = out.Bytes()

	return im, nil
}

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(r io.Reader, contentType string) (*imageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImagePng:
		i, err = StrippedPngDecode(r)
		if err != nil {
			return nil, err
		}
	case mimeImageGif:
		i, err = gif.Decode(r)
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
	return &imageMeta{
		small: out.Bytes(),
	}, nil
}
