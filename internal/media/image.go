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

type imageAndMeta struct {
	image    []byte
	width    int
	height   int
	size     int
	aspect   float64
	blurhash string
}

func (m *manager) processImage(ctx context.Context, data []byte, contentType string, accountID string) {

	var clean []byte
	var err error
	var original *imageAndMeta
	var small *imageAndMeta

	switch contentType {
	case mimeImageJpeg, mimeImagePng:
		// first 'clean' image by purging exif data from it
		var exifErr error
		if clean, exifErr = purgeExif(data); exifErr != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", exifErr)
		}
		original, err = decodeImage(clean, contentType)
	case mimeImageGif:
		// gifs are already clean - no exif data to remove
		clean = data
		original, err = decodeGif(clean)
	default:
		err = fmt.Errorf("content type %s not a processible image type", contentType)
	}

	if err != nil {
		return nil, err
	}

	small, err = deriveThumbnail(clean, contentType, thumbnailMaxWidth, thumbnailMaxHeight)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new id for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	attachmentID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	originalURL := uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeOriginal), attachmentID, extension)
	smallURL := uris.GenerateURIForAttachment(accountID, string(TypeAttachment), string(SizeSmall), attachmentID, "jpeg") // all thumbnails/smalls are encoded as jpeg

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s.%s", accountID, TypeAttachment, SizeOriginal, attachmentID, extension)
	if err := m.storage.Put(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s.jpeg", accountID, TypeAttachment, SizeSmall, attachmentID) // all thumbnails/smalls are encoded as jpeg
	if err := m.storage.Put(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	attachment := &gtsmodel.MediaAttachment{
		ID:        attachmentID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: "",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		Type:      gtsmodel.FileTypeImage,
		FileMeta: gtsmodel.FileMeta{
			Original: gtsmodel.Original{
				Width:  original.width,
				Height: original.height,
				Size:   original.size,
				Aspect: original.aspect,
			},
			Small: gtsmodel.Small{
				Width:  small.width,
				Height: small.height,
				Size:   small.size,
				Aspect: small.aspect,
			},
		},
		AccountID:         accountID,
		Description:       "",
		ScheduledStatusID: "",
		Blurhash:          small.blurhash,
		Processing:        2,
		File: gtsmodel.File{
			Path:        originalPath,
			ContentType: contentType,
			FileSize:    len(original.image),
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        smallPath,
			ContentType: mimeJpeg, // all thumbnails/smalls are encoded as jpeg
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			URL:         smallURL,
			RemoteURL:   "",
		},
		Avatar: false,
		Header: false,
	}

	return attachment, nil
}

func decodeGif(b []byte) (*imageAndMeta, error) {
	gif, err := gif.DecodeAll(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	// use the first frame to get the static characteristics
	width := gif.Config.Width
	height := gif.Config.Height
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

func decodeImage(b []byte, contentType string) (*imageAndMeta, error) {
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
	case mimeImageJpeg:
		i, err = jpeg.Decode(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
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
	return &imageAndMeta{
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
