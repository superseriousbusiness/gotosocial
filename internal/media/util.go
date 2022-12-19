/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"io"

	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var SupportedMIMETypes = []string{
	mimeImageJpeg,
	mimeImageGif,
	mimeImagePng,
	mimeImageWebp,
	mimeVideoMp4,
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
