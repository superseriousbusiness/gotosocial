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
	"context"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"
	"time"

	"github.com/buckket/go-blurhash"
	"github.com/nfnt/resize"
	"github.com/superseriousbusiness/exifremove/pkg/exifremove"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	thumbnailMaxWidth  = 512
	thumbnailMaxHeight = 512
)

type ImageMeta struct {
	image    []byte
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string
}

func (m *manager) preProcessImage(ctx context.Context, data []byte, contentType string, accountID string, ai *AdditionalInfo) (*Processing, error) {
	if !supportedImage(contentType) {
		return nil, fmt.Errorf("image type %s not supported", contentType)
	}

	if len(data) == 0 {
		return nil, errors.New("image was of size 0")
	}

	id, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	extension := strings.Split(contentType, "/")[1]

	// populate initial fields on the media attachment -- some of these will be overwritten as we proceed
	attachment := &gtsmodel.MediaAttachment{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		StatusID:  "",
		URL:       uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeOriginal), id, extension),
		RemoteURL: "",
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Focus: gtsmodel.Focus{
				X: 0,
				Y: 0,
			},
		},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          "",
		Processing:        gtsmodel.ProcessingStatusReceived,
		File: gtsmodel.File{
			Path:        fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeOriginal, id, extension),
			ContentType: contentType,
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			URL:         uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeSmall), id, mimeJpeg), // all thumbnails are encoded as jpeg,
			Path:        fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeSmall, id, mimeJpeg),                 // all thumbnails are encoded as jpeg,
			ContentType: mimeJpeg,
			UpdatedAt:   time.Now(),
		},
		Avatar: false,
		Header: false,
	}

	// check if we have additional info to add to the attachment,
	// and overwrite some of the attachment fields if so
	if ai != nil {
		if ai.CreatedAt != nil {
			attachment.CreatedAt = *ai.CreatedAt
		}

		if ai.StatusID != nil {
			attachment.StatusID = *ai.StatusID
		}

		if ai.RemoteURL != nil {
			attachment.RemoteURL = *ai.RemoteURL
		}

		if ai.Description != nil {
			attachment.Description = *ai.Description
		}

		if ai.ScheduledStatusID != nil {
			attachment.ScheduledStatusID = *ai.ScheduledStatusID
		}

		if ai.Blurhash != nil {
			attachment.Blurhash = *ai.Blurhash
		}

		if ai.Avatar != nil {
			attachment.Avatar = *ai.Avatar
		}

		if ai.Header != nil {
			attachment.Header = *ai.Header
		}

		if ai.FocusX != nil {
			attachment.FileMeta.Focus.X = *ai.FocusX
		}

		if ai.FocusY != nil {
			attachment.FileMeta.Focus.Y = *ai.FocusY
		}
	}

	media := &Processing{
		attachment:    attachment,
		rawData:       data,
		thumbstate:    received,
		fullSizeState: received,
		database:      m.db,
		storage:       m.storage,
	}

	return media, nil
}

func decodeGif(b []byte) (*ImageMeta, error) {
	gif, err := gif.DecodeAll(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// use the first frame to get the static characteristics
	width := gif.Config.Width
	height := gif.Config.Height
	size := width * height
	aspect := float64(width) / float64(height)

	return &ImageMeta{
		image:  b,
		width:  width,
		height: height,
		size:   size,
		aspect: aspect,
	}, nil
}

func decodeImage(b []byte, contentType string) (*ImageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
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

	return &ImageMeta{
		image:  b,
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
func deriveThumbnail(b []byte, contentType string, createBlurhash bool) (*ImageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImageJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
	case mimeImageGif:
		i, err = gif.Decode(bytes.NewReader(b))
	default:
		err = fmt.Errorf("content type %s can't be thumbnailed", contentType)
	}

	if err != nil {
		return nil, err
	}

	if i == nil {
		return nil, errors.New("processed image was nil")
	}

	thumb := resize.Thumbnail(thumbnailMaxWidth, thumbnailMaxHeight, i, resize.NearestNeighbor)
	width := thumb.Bounds().Size().X
	height := thumb.Bounds().Size().Y
	size := width * height
	aspect := float64(width) / float64(height)

	im := &ImageMeta{
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
			return nil, err
		}
		im.blurhash = bh
	}

	out := &bytes.Buffer{}
	if err := jpeg.Encode(out, thumb, &jpeg.Options{
		// Quality isn't extremely important for thumbnails, so 75 is "good enough"
		Quality: 75,
	}); err != nil {
		return nil, err
	}

	im.image = out.Bytes()

	return im, nil
}

// deriveStaticEmojji takes a given gif or png of an emoji, decodes it, and re-encodes it as a static png.
func deriveStaticEmoji(b []byte, contentType string) (*ImageMeta, error) {
	var i image.Image
	var err error

	switch contentType {
	case mimeImagePng:
		i, err = png.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	case mimeImageGif:
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
	return &ImageMeta{
		image: out.Bytes(),
	}, nil
}

// purgeExif is a little wrapper for the action of removing exif data from an image.
// Only pass pngs or jpegs to this function.
func purgeExif(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("passed image was not valid")
	}

	clean, err := exifremove.Remove(data)
	if err != nil {
		return nil, fmt.Errorf("could not purge exif from image: %s", err)
	}

	if len(clean) == 0 {
		return nil, errors.New("purged image was not valid")
	}

	return clean, nil
}
