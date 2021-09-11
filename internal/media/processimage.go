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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (mh *mediaHandler) processImageAttachment(data []byte, minAttachment *gtsmodel.MediaAttachment) (*gtsmodel.MediaAttachment, error) {
	var clean []byte
	var err error
	var original *imageAndMeta
	var small *imageAndMeta

	contentType := minAttachment.File.ContentType

	switch contentType {
	case MIMEJpeg, MIMEPng:
		if clean, err = purgeExif(data); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
		original, err = deriveImage(clean, contentType)
		if err != nil {
			return nil, fmt.Errorf("error parsing image: %s", err)
		}
	case MIMEGif:
		clean = data
		original, err = deriveGif(clean, contentType)
		if err != nil {
			return nil, fmt.Errorf("error parsing gif: %s", err)
		}
	default:
		return nil, errors.New("media type unrecognized")
	}

	small, err = deriveThumbnail(clean, contentType, 256, 256)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new id for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/attachment/original/%s.%s", URLbase, minAttachment.AccountID, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/attachment/small/%s.jpeg", URLbase, minAttachment.AccountID, newMediaID) // all thumbnails/smalls are encoded as jpeg

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s.%s", minAttachment.AccountID, Attachment, Original, newMediaID, extension)
	if err := mh.storage.Put(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s.jpeg", minAttachment.AccountID, Attachment, Small, newMediaID) // all thumbnails/smalls are encoded as jpeg
	if err := mh.storage.Put(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	minAttachment.FileMeta.Original = gtsmodel.Original{
		Width:  original.width,
		Height: original.height,
		Size:   original.size,
		Aspect: original.aspect,
	}

	minAttachment.FileMeta.Small = gtsmodel.Small{
		Width:  small.width,
		Height: small.height,
		Size:   small.size,
		Aspect: small.aspect,
	}

	attachment := &gtsmodel.MediaAttachment{
		ID:                newMediaID,
		StatusID:          minAttachment.StatusID,
		URL:               originalURL,
		RemoteURL:         minAttachment.RemoteURL,
		CreatedAt:         minAttachment.CreatedAt,
		UpdatedAt:         minAttachment.UpdatedAt,
		Type:              gtsmodel.FileTypeImage,
		FileMeta:          minAttachment.FileMeta,
		AccountID:         minAttachment.AccountID,
		Description:       minAttachment.Description,
		ScheduledStatusID: minAttachment.ScheduledStatusID,
		Blurhash:          original.blurhash,
		Processing:        2,
		File: gtsmodel.File{
			Path:        originalPath,
			ContentType: contentType,
			FileSize:    len(original.image),
			UpdatedAt:   time.Now(),
		},
		Thumbnail: gtsmodel.Thumbnail{
			Path:        smallPath,
			ContentType: MIMEJpeg, // all thumbnails/smalls are encoded as jpeg
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			URL:         smallURL,
			RemoteURL:   minAttachment.Thumbnail.RemoteURL,
		},
		Avatar: minAttachment.Avatar,
		Header: minAttachment.Header,
	}

	return attachment, nil
}
