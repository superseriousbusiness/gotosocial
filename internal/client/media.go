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

package client

import "mime/multipart"

// MediaRequest represents the form data parameters submitted by a client during a media upload request.
// See: https://docs.joinmastodon.org/methods/statuses/media/
type MediaRequest struct {
	File        *multipart.FileHeader `form:"file"`
	Thumbnail   *multipart.FileHeader `form:"thumbnail"`
	Description string                `form:"description"`
	Focus       string                `form:"focus"`
}

// MediaResponse represents the object returned to a client after a successful media upload request.
// See: https://docs.joinmastodon.org/methods/statuses/media/
type MediaResponse struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	PreviewURL  string    `json:"preview_url"`
	RemoteURL   string    `json:"remote_url"`
	TextURL     string    `json:"text_url"`
	Meta        MediaMeta `json:"meta"`
	Description string    `json:"description"`
	Blurhash    string    `json:"blurhash"`
}

// MediaMeta describes the media that's just been uploaded. It should be returned to the caller as part of MediaResponse.
type MediaMeta struct {
	Focus    MediaFocus      `json:"focus"`
	Original MediaDimensions `json:"original"`
	Small    MediaDimensions `json:"small"`
}

// MediaFocus describes the focal point of a piece of media. It should be returned to the caller as part of MediaMeta.
type MediaFocus struct {
	X float32 `json:"x"` // should be between -1 and 1
	Y float32 `json:"y"` // should be between -1 and 1
}

// MediaDimensions describes the physical properties of a piece of media. It should be returned to the caller as part of MediaMeta.
type MediaDimensions struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Size   string  `json:"size"`
	Aspect float32 `json:"aspect"`
}
