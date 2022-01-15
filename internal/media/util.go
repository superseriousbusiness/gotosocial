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

	"github.com/h2non/filetype"
	"github.com/superseriousbusiness/gotosocial/internal/db"
)

// parseContentType parses the MIME content type from a file, returning it as a string in the form (eg., "image/jpeg").
// Returns an error if the content type is not something we can process.
func parseContentType(content []byte) (string, error) {
	// read in the first bytes of the file
	fileHeader := make([]byte, maxFileHeaderBytes)
	if _, err := bytes.NewReader(content).Read(fileHeader); err != nil {
		return "", fmt.Errorf("could not read first magic bytes of file: %s", err)
	}

	kind, err := filetype.Match(fileHeader)
	if err != nil {
		return "", err
	}

	if kind == filetype.Unknown {
		return "", errors.New("filetype unknown")
	}

	return kind.MIME.Value, nil
}

// supportedImage checks mime type of an image against a slice of accepted types,
// and returns True if the mime type is accepted.
func supportedImage(mimeType string) bool {
	acceptedImageTypes := []string{
		mimeImageJpeg,
		mimeImageGif,
		mimeImagePng,
	}
	for _, accepted := range acceptedImageTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// supportedEmoji checks that the content type is image/png -- the only type supported for emoji.
func supportedEmoji(mimeType string) bool {
	acceptedEmojiTypes := []string{
		mimeImageGif,
		mimeImagePng,
	}
	for _, accepted := range acceptedEmojiTypes {
		if mimeType == accepted {
			return true
		}
	}
	return false
}

// ParseMediaType converts s to a recognized MediaType, or returns an error if unrecognized
func ParseMediaType(s string) (Type, error) {
	switch s {
	case string(TypeAttachment):
		return TypeAttachment, nil
	case string(TypeHeader):
		return TypeHeader, nil
	case string(TypeAvatar):
		return TypeAvatar, nil
	case string(TypeEmoji):
		return TypeEmoji, nil
	}
	return "", fmt.Errorf("%s not a recognized MediaType", s)
}

// ParseMediaSize converts s to a recognized MediaSize, or returns an error if unrecognized
func ParseMediaSize(s string) (Size, error) {
	switch s {
	case string(SizeSmall):
		return SizeSmall, nil
	case string(SizeOriginal):
		return SizeOriginal, nil
	case string(SizeStatic):
		return SizeStatic, nil
	}
	return "", fmt.Errorf("%s not a recognized MediaSize", s)
}

// putOrUpdate is just a convenience function for first trying to PUT the attachment or emoji in the database,
// and then if that doesn't work because the attachment/emoji already exists, updating it instead.
func putOrUpdate(ctx context.Context, database db.DB, i interface{}) error {
	if err := database.Put(ctx, i); err != nil {
		if err != db.ErrAlreadyExists {
			return fmt.Errorf("putOrUpdate: proper error while putting: %s", err)
		}
		if err := database.UpdateByPrimaryKey(ctx, i); err != nil {
			return fmt.Errorf("putOrUpdate: error while updating: %s", err)
		}
	}

	return nil
}
