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
	"errors"
	"fmt"

	"github.com/h2non/filetype"
	"github.com/sirupsen/logrus"
)

// parseContentType parses the MIME content type from a file, returning it as a string in the form (eg., "image/jpeg").
// Returns an error if the content type is not something we can process.
//
// Fileheader should be no longer than 262 bytes; anything more than this is inefficient.
func parseContentType(fileHeader []byte) (string, error) {
	if fhLength := len(fileHeader); fhLength > maxFileHeaderBytes {
		return "", fmt.Errorf("parseContentType requires %d bytes max, we got %d", maxFileHeaderBytes, fhLength)
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

// supportedEmoji checks that the content type is image/png or image/gif -- the only types supported for emoji.
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

// logrusWrapper is just a util for passing the logrus logger into the cron logging system.
type logrusWrapper struct {
}

// Info logs routine messages about cron's operation.
func (l *logrusWrapper) Info(msg string, keysAndValues ...interface{}) {
	logrus.Info("media manager cron logger: ", msg, keysAndValues)
}

// Error logs an error condition.
func (l *logrusWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	logrus.Error("media manager cron logger: ", err, msg, keysAndValues)
}
