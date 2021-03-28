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
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
	exifremove "github.com/scottleedavis/go-exif-remove"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

var (
	acceptedImageTypes = []string{
		"jpeg",
		"gif",
		"png",
	}
)

// MediaHandler provides an interface for parsing, storing, and retrieving media objects like photos, videos, and gifs.
type MediaHandler interface {
	// SetHeaderForAccountID takes a new header image for an account, checks it out, removes exif data from it,
	// puts it in whatever storage backend we're using, sets the relevant fields in the database for the new image,
	// and then returns information to the caller about the new header's web location.
	SetHeaderForAccountID(f multipart.File, id string) (*HeaderInfo, error)
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

func (mh *mediaHandler) SetHeaderForAccountID(f multipart.File, accountID string) (*HeaderInfo, error) {
	l := mh.log.WithField("func", "SetHeaderForAccountID")

	// make sure we can handle this
	extension, err := processableHeaderOrAvi(f)
	if err != nil {
		return nil, err
	}

	// extract the bytes
	imageBytes := []byte{}
	size, err := f.Read(imageBytes)
	if err != nil {
		return nil, fmt.Errorf("error reading file bytes: %s", err)
	}
	l.Tracef("read %d bytes of file", size)

	// close the open file--we don't need it anymore now we have the bytes
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("error closing file: %s", err)
	}

	// remove exif data from images because fuck that shit
	cleanBytes := []byte{}
	if extension == "jpeg" || extension == "png" {
		cleanBytes, err = exifremove.Remove(imageBytes)
		if err != nil {
			return nil, fmt.Errorf("error removing exif from image: %s", err)
		}
	} else {
		// our only other accepted image type (gif) doesn't need cleaning
		cleanBytes = imageBytes
	}

	// now put it in storage, take a new uuid for the name of the file so we don't store any unnecessary info about it
	path := fmt.Sprintf("/%s/media/headers/%s.%s", accountID, uuid.NewString(), extension)
	if err := mh.storage.StoreFileAt(path, cleanBytes); err != nil {
		return nil, fmt.Errorf("storage error: %s", err)
	}

	return nil, nil
}

func processableHeaderOrAvi(f multipart.File) (string, error) {
	extension := ""

	head := make([]byte, 261)
	_, err := f.Read(head)
	if err != nil {
		return extension, fmt.Errorf("could not read first magic bytes of file: %s", err)
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return extension, err
	}

	if kind == filetype.Unknown || !filetype.IsImage(head) {
		return extension, errors.New("filetype is not an image")
	}

	if !supportedImageType(kind.MIME.Subtype) {
		return extension, fmt.Errorf("%s is not an accepted image type", kind.MIME.Value)
	}

	extension = kind.MIME.Subtype

	return extension, nil
}

func supportedImageType(have string) bool {
	for _, accepted := range acceptedImageTypes {
		if have == accepted {
			return true
		}
	}
	return false
}
