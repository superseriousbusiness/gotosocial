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
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/h2non/filetype"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// AllSupportedMIMETypes just returns all media
// MIME types supported by this instance.
func AllSupportedMIMETypes() []string {
	return []string{
		mimeImageJpeg,
		mimeImageGif,
		mimeImagePng,
		mimeImageWebp,
		mimeVideoMp4,
	}
}

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

// supportedAttachment checks mime type of an attachment against a
// slice of accepted types, and returns True if the mime type is accepted.
func supportedAttachment(mimeType string) bool {
	for _, accepted := range AllSupportedMIMETypes() {
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
type logrusWrapper struct{}

// Info logs routine messages about cron's operation.
func (l *logrusWrapper) Info(msg string, keysAndValues ...interface{}) {
	log.Info("media manager cron logger: ", msg, keysAndValues)
}

// Error logs an error condition.
func (l *logrusWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Error("media manager cron logger: ", err, msg, keysAndValues)
}

// lengthReader wraps a reader and reads the length of total bytes written as it goes.
type lengthReader struct {
	source io.Reader
	length int64
}

func (r *lengthReader) Read(b []byte) (int, error) {
	n, err := r.source.Read(b)
	r.length += int64(n)
	return n, err
}

// putStream either puts a file with a known fileSize into storage directly, and returns the
// fileSize unchanged, or it wraps the reader with a lengthReader and returns the discovered
// fileSize.
func putStream(ctx context.Context, storage *storage.Driver, key string, r io.Reader, fileSize int64) (int64, error) {
	if fileSize > 0 {
		return fileSize, storage.PutStream(ctx, key, r)
	}

	lr := &lengthReader{
		source: r,
	}

	err := storage.PutStream(ctx, key, lr)
	return lr.length, err
}
