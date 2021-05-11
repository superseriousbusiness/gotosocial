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

package model

import "mime/multipart"

// AttachmentRequest represents the form data parameters submitted by a client during a media upload request.
// See: https://docs.joinmastodon.org/methods/statuses/media/
type AttachmentRequest struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	Description string                `form:"description"`
	Focus       string                `form:"focus"`
}

// AttachmentRequest represents the form data parameters submitted by a client during a media update/PUT request.
// See: https://docs.joinmastodon.org/methods/statuses/media/
type AttachmentUpdateRequest struct {
	Description *string                `form:"description" json:"description" xml:"description"`
	Focus       *string                `form:"focus" json:"focus" xml:"focus"`
}

// Attachment represents the object returned to a client after a successful media upload request.
// See: https://docs.joinmastodon.org/methods/statuses/media/
type Attachment struct {
	// The ID of the attachment in the database.
	ID string `json:"id"`
	// The type of the attachment.
	// 	unknown = unsupported or unrecognized file type.
	// 	image = Static image.
	// 	gifv = Looping, soundless animation.
	// 	video = Video clip.
	// 	audio = Audio track.
	Type string `json:"type"`
	// The location of the original full-size attachment.
	URL string `json:"url"`
	// The location of a scaled-down preview of the attachment.
	PreviewURL string `json:"preview_url"`
	// The location of the full-size original attachment on the remote server.
	RemoteURL string `json:"remote_url,omitempty"`
	// The location of a scaled-down preview of the attachment on the remote server.
	PreviewRemoteURL string `json:"preview_remote_url,omitempty"`
	// A shorter URL for the attachment.
	TextURL string `json:"text_url,omitempty"`
	// Metadata returned by Paperclip.
	// May contain subtrees small and original, as well as various other top-level properties.
	// More importantly, there may be another top-level focus Hash object as of 2.3.0, with coordinates can be used for smart thumbnail cropping.
	// See https://docs.joinmastodon.org/methods/statuses/media/#focal-points points for more.
	Meta MediaMeta `json:"meta,omitempty"`
	// Alternate text that describes what is in the media attachment, to be used for the visually impaired or when media attachments do not load.
	Description string `json:"description,omitempty"`
	// A hash computed by the BlurHash algorithm, for generating colorful preview thumbnails when media has not been downloaded yet.
	// See https://github.com/woltapp/blurhash
	Blurhash string `json:"blurhash,omitempty"`
}

// MediaMeta describes the returned media
type MediaMeta struct {
	Length        string          `json:"length,omitempty"`
	Duration      float32         `json:"duration,omitempty"`
	FPS           uint16          `json:"fps,omitempty"`
	Size          string          `json:"size,omitempty"`
	Width         int             `json:"width,omitempty"`
	Height        int             `json:"height,omitempty"`
	Aspect        float32         `json:"aspect,omitempty"`
	AudioEncode   string          `json:"audio_encode,omitempty"`
	AudioBitrate  string          `json:"audio_bitrate,omitempty"`
	AudioChannels string          `json:"audio_channels,omitempty"`
	Original      MediaDimensions `json:"original"`
	Small         MediaDimensions `json:"small,omitempty"`
	Focus         MediaFocus      `json:"focus,omitempty"`
}

// MediaFocus describes the focal point of a piece of media. It should be returned to the caller as part of MediaMeta.
type MediaFocus struct {
	X float32 `json:"x"` // should be between -1 and 1
	Y float32 `json:"y"` // should be between -1 and 1
}

// MediaDimensions describes the physical properties of a piece of media. It should be returned to the caller as part of MediaMeta.
type MediaDimensions struct {
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	FrameRate string  `json:"frame_rate,omitempty"`
	Duration  float32 `json:"duration,omitempty"`
	Bitrate   int     `json:"bitrate,omitempty"`
	Size      string  `json:"size,omitempty"`
	Aspect    float32 `json:"aspect,omitempty"`
}
