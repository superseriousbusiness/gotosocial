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
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/buckket/go-blurhash"
	"github.com/disintegration/imaging"
)

type mediaMeta struct {
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string // defined only for calls to deriveThumbnail if createBlurhash is true
	small    []byte // defined only for calls to deriveStaticEmoji or deriveThumbnail
}

// deriveThumbnail returns a byte slice and metadata for a thumbnail
// of a given piece of media, or an error if something goes wrong.
//
// If createBlurhash is true, then a blurhash will also be generated from a tiny
// version of the image. This costs precious CPU cycles, so only use it if you
// really need a blurhash and don't have one already.
//
// If createBlurhash is false, then the blurhash field on the returned ImageAndMeta
// will be an empty string.
func deriveThumbnail(r io.Reader, contentType string, createBlurhash bool) (*mediaMeta, error) {
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
	case mimeVideoMp4:
		i, err = extractFromVideo(r)
	default:
		err = fmt.Errorf("content type %s can't be thumbnailed", contentType)
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
	aspect := float64(thumbX) / float64(thumbY)

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
