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
	"io"
	"time"
)

// maxFileHeaderBytes represents the maximum amount of bytes we want
// to examine from the beginning of a file to determine its type.
//
// See: https://en.wikipedia.org/wiki/File_format#File_header
// and https://github.com/h2non/filetype
const maxFileHeaderBytes = 261

// mime consts
const (
	mimeImage = "image"
	mimeVideo = "video"

	mimeJpeg      = "jpeg"
	mimeImageJpeg = mimeImage + "/" + mimeJpeg

	mimeGif      = "gif"
	mimeImageGif = mimeImage + "/" + mimeGif

	mimePng      = "png"
	mimeImagePng = mimeImage + "/" + mimePng

	mimeWebp      = "webp"
	mimeImageWebp = mimeImage + "/" + mimeWebp

	mimeMp4      = "mp4"
	mimeVideoMp4 = mimeVideo + "/" + mimeMp4
)

type processState int32

const (
	received processState = iota // processing order has been received but not done yet
	complete                     // processing order has been completed successfully
	errored                      // processing order has been completed with an error
)

// EmojiMaxBytes is the maximum permitted bytes of an emoji upload (50kb)
// const EmojiMaxBytes = 51200

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

// AdditionalEmojiInfo represents additional information
// that should be taken into account when processing an emoji.
type AdditionalEmojiInfo struct {
	// Time that this emoji was created; defaults to time.Now().
	CreatedAt *time.Time
	// Domain the emoji originated from. Blank for this instance's domain. Defaults to "".
	Domain *string
	// URL of this emoji on a remote instance; defaults to "".
	ImageRemoteURL *string
	// URL of the static version of this emoji on a remote instance; defaults to "".
	ImageStaticRemoteURL *string
	// Whether this emoji should be disabled (not shown) on this instance; defaults to false.
	Disabled *bool
	// Whether this emoji should be visible in the instance's emoji picker; defaults to true.
	VisibleInPicker *bool
	// ID of the category this emoji should be placed in; defaults to "".
	CategoryID *string
}

// DataFunc represents a function used to retrieve the raw bytes of a piece of media.
type DataFunc func(ctx context.Context) (reader io.ReadCloser, fileSize int64, err error)

// PostDataCallbackFunc represents a function executed after the DataFunc has been executed,
// and the returned reader has been read. It can be used to clean up any remaining resources.
//
// This can be set to nil, and will then not be executed.
type PostDataCallbackFunc func(ctx context.Context) error

type mediaMeta struct {
	width    int
	height   int
	size     int
	aspect   float32
	blurhash string
	small    []byte

	// video-specific properties
	duration  float32
	framerate float32
	bitrate   uint64
}
