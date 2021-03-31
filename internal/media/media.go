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
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// MediaHandler provides an interface for parsing, storing, and retrieving media objects like photos, videos, and gifs.
type MediaHandler interface {
	// SetHeaderOrAvatarForAccountID takes a new header image for an account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
	// and then returns information to the caller about the new header.
	SetHeaderOrAvatarForAccountID(img []byte, accountID string, headerOrAvi string) (*model.MediaAttachment, error)
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

func (mh *mediaHandler) SetHeaderOrAvatarForAccountID(img []byte, accountID string, headerOrAvi string) (*model.MediaAttachment, error) {
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

/*
	HELPER FUNCTIONS
*/

func (mh *mediaHandler) processHeaderOrAvi(imageBytes []byte, contentType string, headerOrAvi string, accountID string) (*model.MediaAttachment, error) {
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

	clean := []byte{}
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

	base := fmt.Sprintf("%s://%s%s", mh.config.StorageConfig.ServeProtocol, mh.config.StorageConfig.ServeHost, mh.config.StorageConfig.ServeBasePath, )

	// we store the original...
	originalPath := fmt.Sprintf("%s/%s/%s/original/%s.%s", base, accountID, headerOrAvi, newMediaID, extension)
	if err := mh.storage.StoreFileAt(originalPath, original.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}
	// and a thumbnail...
	smallPath := fmt.Sprintf("%s/%s/%s/small/%s.%s", base, accountID, headerOrAvi, newMediaID, extension)
	if err := mh.storage.StoreFileAt(smallPath, small.image); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	ma := &model.MediaAttachment{
		ID:        newMediaID,
		StatusID:  "",
		RemoteURL: "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Type:      model.FileTypeImage,
		FileMeta: model.FileMeta{
			Original: model.Original{
				Width:  original.width,
				Height: original.height,
				Size:   original.size,
				Aspect: original.aspect,
			},
			Small: model.Small{
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
		File: model.File{
			Path:        originalPath,
			ContentType: contentType,
			FileSize:    len(original.image),
			UpdatedAt:   time.Now(),
		},
		Thumbnail: model.Thumbnail{
			Path:        smallPath,
			ContentType: contentType,
			FileSize:    len(small.image),
			UpdatedAt:   time.Now(),
			RemoteURL:   "",
		},
		Avatar: isAvatar,
		Header: isHeader,
	}

	return ma, nil
}
