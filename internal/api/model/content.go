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

package model

import (
	"io"

	"code.superseriousbusiness.org/gotosocial/internal/storage"
)

// Content wraps everything needed to serve a blob of content (some kind of media) through the API.
type Content struct {
	// MIME content type
	ContentType string
	// ContentLength in bytes
	ContentLength int64
	// Actual content
	Content io.ReadCloser
	// Resource URL to forward to if the file can be fetched from the storage directly (e.g signed S3 URL)
	URL *storage.PresignedURL
}

// GetContentRequestForm describes a piece of content desired by the caller of the fileserver API.
type GetContentRequestForm struct {
	// AccountID of the content owner
	AccountID string
	// MediaType of the content (should be convertible to a media.MediaType)
	MediaType string
	// MediaSize of the content (should be convertible to a media.MediaSize)
	MediaSize string
	// Filename of the content
	FileName string
}
