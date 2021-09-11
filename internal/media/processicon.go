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

func (mh *mediaHandler) processHeaderOrAvi(imageBytes []byte, contentType string, mediaType Type, accountID string, remoteURL string) (*gtsmodel.MediaAttachment, error) {
	var isHeader bool
	var isAvatar bool

	switch mediaType {
	case Header:
		isHeader = true
	case Avatar:
		isAvatar = true
	default:
		return nil, errors.New("header or avatar not selected")
	}

	var clean []byte
	var err error

	var original *imageAndMeta
	switch contentType {
	case MIMEJpeg:
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
		original, err = deriveImage(clean, contentType)
	case MIMEPng:
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
		original, err = deriveImage(clean, contentType)
	case MIMEGif:
		clean = imageBytes
		original, err = deriveGif(clean, contentType)
	default:
		return nil, errors.New("media type unrecognized")
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing image: %s", err)
	}

	small, err := deriveThumbnail(clean, contentType, 256, 256)
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
	originalURL := fmt.Sprintf("%s/%s/%s/original/%s.%s", URLbase, accountID, mediaType, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/%s/small/%s.%s", URLbase, accountID, mediaType, newMediaID, extension)

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/%s.%s", accountID, mediaType, Original, newMediaID, extension)
	if err := mh.storage.Put(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/%s.%s", accountID, mediaType, Small, newMediaID, extension)
	if err := mh.storage.Put(smallPath, small.image); err != nil {
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
			ContentType: contentType,
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			URL:         smallURL,
			RemoteURL:   "",
		},
		Avatar: isAvatar,
		Header: isHeader,
	}

	return ma, nil
}
