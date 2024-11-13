// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package media

import (
	"context"
	"io"
)

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

// AdditionalMediaInfo represents additional information that
// should be added to attachment when processing a piece of media.
type AdditionalMediaInfo struct {

	// ID of the status to which this
	// media is attached; defaults to "".
	StatusID *string

	// URL of the media on a
	// remote instance; defaults to "".
	RemoteURL *string

	// Image description of
	// this media; defaults to "".
	Description *string

	// Blurhash of this
	// media; defaults to "".
	Blurhash *string

	// ID of the scheduled status to which
	// this media is attached; defaults to "".
	ScheduledStatusID *string

	// Mark this media as in-use
	// as an avatar; defaults to false.
	Avatar *bool

	// Mark this media as in-use
	// as a header; defaults to false.
	Header *bool

	// X focus coordinate for
	// this media; defaults to 0.
	FocusX *float32

	// Y focus coordinate for
	// this media; defaults to 0.
	FocusY *float32
}

// AdditionalEmojiInfo represents additional information
// that should be taken into account when processing an emoji.
type AdditionalEmojiInfo struct {

	// ActivityPub URI of
	// this remote emoji.
	URI *string

	// Domain the emoji originated from. Blank
	// for this instance's domain. Defaults to "".
	Domain *string

	// URL of this emoji on a
	// remote instance; defaults to "".
	ImageRemoteURL *string

	// URL of the static version of this emoji
	// on a remote instance; defaults to "".
	ImageStaticRemoteURL *string

	// Whether this emoji should be disabled (not
	// shown) on this instance; defaults to false.
	Disabled *bool

	// Whether this emoji should be visible in
	// the instance's emoji picker; defaults to true.
	VisibleInPicker *bool

	// ID of the category this emoji
	// should be placed in; defaults to "".
	CategoryID *string
}

// DataFunc represents a function used to retrieve the raw bytes of a piece of media.
type DataFunc func(ctx context.Context) (reader io.ReadCloser, err error)
