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
	"time"

	"github.com/h2non/filetype"
)

// mime consts
const (
	mimeImage = "image"

	mimeJpeg      = "jpeg"
	mimeImageJpeg = mimeImage + "/" + mimeJpeg

	mimeGif      = "gif"
	mimeImageGif = mimeImage + "/" + mimeGif

	mimePng      = "png"
	mimeImagePng = mimeImage + "/" + mimePng
)

// EmojiMaxBytes is the maximum permitted bytes of an emoji upload (50kb)
// const EmojiMaxBytes = 51200

// maxFileHeaderBytes represents the maximum amount of bytes we want
// to examine from the beginning of a file to determine its type.
//
// See: https://en.wikipedia.org/wiki/File_format#File_header
// and https://github.com/h2non/filetype
const maxFileHeaderBytes = 262

type Size string

const (
	SizeSmall    Size = "small"    // SizeSmall is the key for small/thumbnail versions of media
	SizeOriginal Size = "original" // SizeOriginal is the key for original/fullsize versions of media and emoji
	SizeStatic   Size = "static"   // SizeStatic is the key for static (non-animated) versions of emoji
)

type Type string

const (
	TypeAttachment Type = "attachment" // TypeAttachment is the key for media attachments
	TypeHeader     Type = "header"     // TypeHeader is the key for profile header requests
	TypeAvatar     Type = "avatar"     // TypeAvatar is the key for profile avatar requests
	TypeEmoji      Type = "emoji"      // TypeEmoji is the key for emoji type requests
)

// AdditionalMediaInfo represents additional information that should be added to an attachment
// when processing a piece of media.
type AdditionalMediaInfo struct {
	// Time that this media was created; defaults to time.Now().
	CreatedAt *time.Time
	// ID of the status to which this media is attached; defaults to "".
	StatusID *string
	// URL of the media on a remote instance; defaults to "".
	RemoteURL *string
	// Image description of this media; defaults to "".
	Description *string
	// Blurhash of this media; defaults to "".
	Blurhash *string
	// ID of the scheduled status to which this media is attached; defaults to "".
	ScheduledStatusID *string
	// Mark this media as in-use as an avatar; defaults to false.
	Avatar *bool
	// Mark this media as in-use as a header; defaults to false.
	Header *bool
	// X focus coordinate for this media; defaults to 0.
	FocusX *float32
	// Y focus coordinate for this media; defaults to 0.
	FocusY *float32
}

type AdditionalEmojiInfo struct {
	
}

// DataFunc represents a function used to retrieve the raw bytes of a piece of media.
type DataFunc func(ctx context.Context) ([]byte, error)

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
