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

	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (mh *mediaHandler) processImageAttachment(data []byte, accountID string, contentType string, remoteURL string) (*gtsmodel.MediaAttachment, error) {
	var clean []byte
	var err error
	var original *imageAndMeta
	var small *imageAndMeta

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

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID := uuid.NewString()

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/attachment/original/%s.%s", URLbase, accountID, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/attachment/small/%s.jpeg", URLbase, accountID, newMediaID) // all thumbnails/smalls are encoded as jpeg

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s/%s.%s", mh.config.StorageConfig.BasePath, accountID, Attachment, Original, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s/%s.jpeg", mh.config.StorageConfig.BasePath, accountID, Attachment, Small, newMediaID) // all thumbnails/smalls are encoded as jpeg
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &gtsmodel.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: remoteURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
			RemoteURL:   "",
		},
		Avatar: false,
		Header: false,
	}

	return ma, nil

}
