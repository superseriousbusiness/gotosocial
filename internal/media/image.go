/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // blank import to support WebP decoding
)

const (
	thumbnailMaxWidth  = 512
	thumbnailMaxHeight = 512
)

func decodeGif(r io.Reader) (*mediaMeta, error) {
	gif, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}

	// use the first frame to get the static characteristics
	width := gif.Config.Width
	height := gif.Config.Height
	size := width * height
	aspect := float32(width) / float32(height)

	return &mediaMeta{
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

func decodeImage(r io.Reader, contentType string) (*mediaMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg, mimeImageWebp:
		i, err = imaging.Decode(r, imaging.AutoOrientation(true))
	case mimeImagePng:
		strippedPngReader := io.Reader(&PNGAncillaryChunkStripper{
			Reader: r,
		})
		i, err = imaging.Decode(strippedPngReader, imaging.AutoOrientation(true))
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
	aspect := float32(width) / float32(height)

	return &mediaMeta{
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(r io.Reader, contentType string) (*mediaMeta, error) {
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
	return &mediaMeta{
		small: out.Bytes(),
	}, nil
}

// deriveThumbnailFromImage returns a byte slice and metadata for a thumbnail
// of a given piece of media, or an error if something goes wrong.
//
// If createBlurhash is true, then a blurhash will also be generated from a tiny
// version of the image. This costs precious CPU cycles, so only use it if you
// really need a blurhash and don't have one already.
//
// If createBlurhash is false, then the blurhash field on the returned ImageAndMeta
// will be an empty string.
func deriveThumbnailFromImage(r io.Reader, contentType string, createBlurhash bool) (*mediaMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg, mimeImageGif, mimeImageWebp:
		i, err = imaging.Decode(r, imaging.AutoOrientation(true))
	case mimeImagePng:
		strippedPngReader := io.Reader(&PNGAncillaryChunkStripper{
			Reader: r,
		})
		i, err = imaging.Decode(strippedPngReader, imaging.AutoOrientation(true))
	default:
		err = fmt.Errorf("content type %s can't be thumbnailed as an image", contentType)
	}

	if err != nil {
		return nil, fmt.Errorf("error decoding %s: %s", contentType, err)
	}

	originalX := i.Bounds().Size().X
	originalY := i.Bounds().Size().Y

	var thumb image.Image
	if originalX <= thumbnailMaxWidth && originalY <= thumbnailMaxHeight {
		// it's already small, no need to resize
		thumb = i
	} else {
		thumb = imaging.Fit(i, thumbnailMaxWidth, thumbnailMaxHeight, imaging.Linear)
	}

	thumbX := thumb.Bounds().Size().X
	thumbY := thumb.Bounds().Size().Y
	size := thumbX * thumbY
	aspect := float32(thumbX) / float32(thumbY)

	im := &mediaMeta{
		width:  thumbX,
		height: thumbY,
		size:   size,
		aspect: aspect,
	}

	if createBlurhash {
		// for generating blurhashes, it's more cost effective to lose detail rather than
		// pass a big image into the blurhash algorithm, so make a teeny tiny version
		tiny := imaging.Resize(thumb, 32, 0, imaging.NearestNeighbor)
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
