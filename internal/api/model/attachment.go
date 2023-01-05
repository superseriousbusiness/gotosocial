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

package model

import "mime/multipart"

// AttachmentRequest models media attachment creation parameters.
//
// swagger: ignore
type AttachmentRequest struct {
	// Media file.
	File *multipart.FileHeader `form:"file" binding:"required"`
	// Description of the media file. Optional.
	// This will be used as alt-text for users of screenreaders etc.
	// example: This is an image of some kittens, they are very cute and fluffy.
	Description string `form:"description"`
	// Focus of the media file. Optional.
	// If present, it should be in the form of two comma-separated floats between -1 and 1.
	// example: -0.5,0.565
	Focus string `form:"focus"`
}

// AttachmentUpdateRequest models an update request for an attachment.
//
// swagger:ignore
type AttachmentUpdateRequest struct {
	// Description of the media file.
	// This will be used as alt-text for users of screenreaders etc.
	// allowEmptyValue: true
	Description *string `form:"description" json:"description" xml:"description"`
	// Focus of the media file.
	// If present, it should be in the form of two comma-separated floats between -1 and 1.
	// allowEmptyValue: true
	Focus *string `form:"focus" json:"focus" xml:"focus"`
}

// Attachment models a media attachment.
//
// swagger:model attachment
type Attachment struct {
	// The ID of the attachment.
	// example: 01FC31DZT1AYWDZ8XTCRWRBYRK
	ID string `json:"id"`
	// The type of the attachment.
	// enum:
	//   - unknown
	//   - image
	//   - gifv
	//   - video
	//   - audio
	// example: image
	Type string `json:"type"`
	// The location of the original full-size attachment.
	// example: https://example.org/fileserver/some_id/attachments/some_id/original/attachment.jpeg
	URL *string `json:"url"`
	// A shorter URL for the attachment.
	// In our case, we just give the URL again since we don't create smaller URLs.
	TextURL string `json:"text_url"`
	// The location of a scaled-down preview of the attachment.
	// example: https://example.org/fileserver/some_id/attachments/some_id/small/attachment.jpeg
	PreviewURL string `json:"preview_url"`
	// The location of the full-size original attachment on the remote server.
	// Only defined for instances other than our own.
	// example: https://some-other-server.org/attachments/original/ahhhhh.jpeg
	RemoteURL *string `json:"remote_url"`
	// The location of a scaled-down preview of the attachment on the remote server.
	// Only defined for instances other than our own.
	// example: https://some-other-server.org/attachments/small/ahhhhh.jpeg
	PreviewRemoteURL *string `json:"preview_remote_url"`
	// Metadata for this attachment.
	Meta MediaMeta `json:"meta,omitempty"`
	// Alt text that describes what is in the media attachment.
	// example: This is a picture of a kitten.
	Description *string `json:"description"`
	// A hash computed by the BlurHash algorithm, for generating colorful preview thumbnails when media has not been downloaded yet.
	// See https://github.com/woltapp/blurhash
	Blurhash string `json:"blurhash,omitempty"`
}

// MediaMeta models media metadata.
// This can be metadata about an image, an audio file, video, etc.
//
// swagger:model mediaMeta
type MediaMeta struct {
	Length string `json:"length,omitempty"`
	// Duration of the media in seconds.
	// Only set for video and audio.
	// example: 5.43
	Duration float32 `json:"duration,omitempty"`
	// Framerate of the media.
	// Only set for video and gifs.
	// example: 30
	FPS uint16 `json:"fps,omitempty"`
	// Size of the media, in the format `[width]x[height]`.
	// Not set for audio.
	// example: 1920x1080
	Size string `json:"size,omitempty"`
	// Width of the media in pixels.
	// Not set for audio.
	// example: 1920
	Width int `json:"width,omitempty"`
	// Height of the media in pixels.
	// Not set for audio.
	// example: 1080
	Height int `json:"height,omitempty"`
	// Aspect ratio of the media.
	// Equal to width / height.
	// example: 1.777777778
	Aspect        float32 `json:"aspect,omitempty"`
	AudioEncode   string  `json:"audio_encode,omitempty"`
	AudioBitrate  string  `json:"audio_bitrate,omitempty"`
	AudioChannels string  `json:"audio_channels,omitempty"`
	// Dimensions of the original media.
	Original MediaDimensions `json:"original"`
	// Dimensions of the thumbnail/small version of the media.
	Small MediaDimensions `json:"small,omitempty"`
	// Focus data for the media.
	Focus MediaFocus `json:"focus,omitempty"`
}

// MediaFocus models the focal point of a piece of media.
//
// swagger:model mediaFocus
type MediaFocus struct {
	// x position of the focus
	// should be between -1 and 1
	X float32 `json:"x"`
	// y position of the focus
	// should be between -1 and 1
	Y float32 `json:"y"`
}

// MediaDimensions models detailed properties of a piece of media.
//
// swagger:model mediaDimensions
type MediaDimensions struct {
	// Width of the media in pixels.
	// Not set for audio.
	// example: 1920
	Width int `json:"width,omitempty"`
	// Height of the media in pixels.
	// Not set for audio.
	// example: 1080
	Height int `json:"height,omitempty"`
	// Framerate of the media.
	// Only set for video and gifs.
	// example: 30
	FrameRate string `json:"frame_rate,omitempty"`
	// Duration of the media in seconds.
	// Only set for video and audio.
	// example: 5.43
	Duration float32 `json:"duration,omitempty"`
	// Bitrate of the media in bits per second.
	// example: 1000000
	Bitrate int `json:"bitrate,omitempty"`
	// Size of the media, in the format `[width]x[height]`.
	// Not set for audio.
	// example: 1920x1080
	Size string `json:"size,omitempty"`
	// Aspect ratio of the media.
	// Equal to width / height.
	// example: 1.777777778
	Aspect float32 `json:"aspect,omitempty"`
}
