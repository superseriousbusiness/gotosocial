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
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// MediaHandler provides an interface for parsing, storing, and retrieving media objects like photos, videos, and gifs.
type MediaHandler interface {
	// SetHeaderOrAvatarForAccountID takes a new header image for an account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
	// and then returns information to the caller about the new header.
	SetHeaderOrAvatarForAccountID(img []byte, accountID string, headerOrAvi string) (*gtsmodel.MediaAttachment, error)

	// ProcessAttachment takes a new attachment and the requesting account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new media,
	// and then returns information to the caller about the attachment.
	ProcessAttachment(img []byte, accountID string) (*gtsmodel.MediaAttachment, error)
}

type mediaHandler struct {
	config  *config.Config
	db      db.DB
	storage storage.Storage
	log     *logrus.Logger
}

func New(config *config.Config, database db.DB, storage storage.Storage, log *logrus.Logger) MediaHandler {
	return &mediaHandler{
		config:  config,
		db:      database,
		storage: storage,
		log:     log,
	}
}

// HeaderInfo wraps the urls at which a Header and a StaticHeader is available from the server.
type HeaderInfo struct {
	// URL to the header
	Header string
	// Static version of the above (eg., a path to a still image if the header is a gif)
	HeaderStatic string
}

/*
	INTERFACE FUNCTIONS
*/

func (mh *mediaHandler) SetHeaderOrAvatarForAccountID(img []byte, accountID string, headerOrAvi string) (*gtsmodel.MediaAttachment, error) {
	l := mh.log.WithField("func", "SetHeaderForAccountID")

	if headerOrAvi != "header" && headerOrAvi != "avatar" {
		return nil, errors.New("header or avatar not selected")
	}

	// make sure we have an image we can handle
	contentType, err := parseContentType(img)
	if err != nil {
		return nil, err
	}
	if !supportedImageType(contentType) {
		return nil, fmt.Errorf("%s is not an accepted image type", contentType)
	}

	if len(img) == 0 {
		return nil, fmt.Errorf("passed reader was of size 0")
	}
	l.Tracef("read %d bytes of file", len(img))

	// process it
	ma, err := mh.processHeaderOrAvi(img, contentType, headerOrAvi, accountID)
	if err != nil {
		return nil, fmt.Errorf("error processing %s: %s", headerOrAvi, err)
	}

	// set it in the database
	if err := mh.db.SetHeaderOrAvatarForAccountID(ma, accountID); err != nil {
		return nil, fmt.Errorf("error putting %s in database: %s", headerOrAvi, err)
	}

	return ma, nil
}

func (mh *mediaHandler) ProcessAttachment(data []byte, accountID string) (*gtsmodel.MediaAttachment, error) {
	contentType, err := parseContentType(data)
	if err != nil {
		return nil, err
	}
	mainType := strings.Split(contentType, "/")[0]
	switch mainType {
	case "video":
		if !supportedVideoType(contentType) {
			return nil, fmt.Errorf("video type %s not supported", contentType)
		}
		if len(data) == 0 {
			return nil, errors.New("video was of size 0")
		}
		if len(data) > mh.config.MediaConfig.MaxVideoSize {
			return nil, fmt.Errorf("video size %d bytes exceeded max video size of %d bytes", len(data), mh.config.MediaConfig.MaxVideoSize)
		}
		return mh.processVideo(data, accountID, contentType)
	case "image":
		if !supportedImageType(contentType) {
			return nil, fmt.Errorf("image type %s not supported", contentType)
		}
		if len(data) == 0 {
			return nil, errors.New("image was of size 0")
		}
		if len(data) > mh.config.MediaConfig.MaxImageSize {
			return nil, fmt.Errorf("image size %d bytes exceeded max image size of %d bytes", len(data), mh.config.MediaConfig.MaxImageSize)
		}
		return mh.processImage(data, accountID, contentType)
	default:
		break
	}
	return nil, fmt.Errorf("content type %s not (yet) supported", contentType)
}

/*
	HELPER FUNCTIONS
*/

func (mh *mediaHandler) processVideo(data []byte, accountID string, contentType string) (*gtsmodel.MediaAttachment, error) {
	return nil, nil
}

func (mh *mediaHandler) processImage(data []byte, accountID string, contentType string) (*gtsmodel.MediaAttachment, error) {
	var clean []byte
	var err error

	switch contentType {
	case "image/jpeg":
		if clean, err = purgeExif(data); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/png":
		if clean, err = purgeExif(data); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/gif":
		clean = data
	default:
		return nil, errors.New("media type unrecognized")
	}

	original, err := deriveImage(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error parsing image: %s", err)
	}

	small, err := deriveThumbnail(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID := uuid.NewString()

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/attachment/original/%s.%s", URLbase, accountID, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/attachment/small/%s.%s", URLbase, accountID, newMediaID, extension)

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/attachment/original/%s.%s", mh.config.StorageConfig.BasePath, accountID, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/attachment/small/%s.%s", mh.config.StorageConfig.BasePath, accountID, newMediaID, extension)
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &gtsmodel.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: "",
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
		Avatar: false,
		Header: false,
	}

	return ma, nil

}

func (mh *mediaHandler) processHeaderOrAvi(imageBytes []byte, contentType string, headerOrAvi string, accountID string) (*gtsmodel.MediaAttachment, error) {
	var isHeader bool
	var isAvatar bool

	switch headerOrAvi {
	case "header":
		isHeader = true
	case "avatar":
		isAvatar = true
	default:
		return nil, errors.New("header or avatar not selected")
	}

	var clean []byte
	var err error

	switch contentType {
	case "image/jpeg":
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/png":
		if clean, err = purgeExif(imageBytes); err != nil {
			return nil, fmt.Errorf("error cleaning exif data: %s", err)
		}
	case "image/gif":
		clean = imageBytes
	default:
		return nil, errors.New("media type unrecognized")
	}

	original, err := deriveImage(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error parsing image: %s", err)
	}

	small, err := deriveThumbnail(clean, contentType)
	if err != nil {
		return nil, fmt.Errorf("error deriving thumbnail: %s", err)
	}

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	extension := strings.Split(contentType, "/")[1]
	newMediaID := uuid.NewString()

	URLbase := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath)
	originalURL := fmt.Sprintf("%s/%s/%s/original/%s.%s", URLbase, accountID, headerOrAvi, newMediaID, extension)
	smallURL := fmt.Sprintf("%s/%s/%s/small/%s.%s", URLbase, accountID, headerOrAvi, newMediaID, extension)

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/original/%s.%s", mh.config.StorageConfig.BasePath, accountID, headerOrAvi, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/small/%s.%s", mh.config.StorageConfig.BasePath, accountID, headerOrAvi, newMediaID, extension)
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &gtsmodel.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		URL:       originalURL,
		RemoteURL: "",
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
