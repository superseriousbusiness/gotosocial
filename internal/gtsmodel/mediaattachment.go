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

package gtsmodel

import (
	"time"
)

// MediaAttachment represents a user-uploaded media attachment: an image/video/audio/gif that is
// somewhere in storage and that can be retrieved and served by the router.
type MediaAttachment struct {
	ID                string           `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt         time.Time        `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	StatusID          string           `bun:"type:CHAR(26),nullzero"`                                      // ID of the status to which this is attached
	URL               string           `bun:",nullzero"`                                                   // Where can the attachment be retrieved on *this* server
	RemoteURL         string           `bun:",nullzero"`                                                   // Where can the attachment be retrieved on a remote server (empty for local media)
	Type              FileType         `bun:",notnull,default:0"`                                          // Type of file (image/gifv/audio/video/unknown)
	FileMeta          FileMeta         `bun:",embed:,notnull"`                                             // Metadata about the file
	AccountID         string           `bun:"type:CHAR(26),nullzero,notnull"`                              // To which account does this attachment belong
	Description       string           `bun:""`                                                            // Description of the attachment (for screenreaders)
	ScheduledStatusID string           `bun:"type:CHAR(26),nullzero"`                                      // To which scheduled status does this attachment belong
	Blurhash          string           `bun:",nullzero"`                                                   // What is the generated blurhash of this attachment
	Processing        ProcessingStatus `bun:",notnull,default:2"`                                          // What is the processing status of this attachment
	File              File             `bun:",embed:file_,notnull,nullzero"`                               // metadata for the whole file
	Thumbnail         Thumbnail        `bun:",embed:thumbnail_,notnull,nullzero"`                          // small image thumbnail derived from a larger image, video, or audio file.
	Avatar            *bool            `bun:",nullzero,notnull,default:false"`                             // Is this attachment being used as an avatar?
	Header            *bool            `bun:",nullzero,notnull,default:false"`                             // Is this attachment being used as a header?
	Cached            *bool            `bun:",nullzero,notnull,default:false"`                             // Is this attachment currently cached by our instance?
}

// IsLocal returns whether media attachment is local.
func (m *MediaAttachment) IsLocal() bool {
	return m.RemoteURL == ""
}

// IsRemote returns whether media attachment is remote.
func (m *MediaAttachment) IsRemote() bool {
	return m.RemoteURL != ""
}

// File refers to the metadata for the whole file
type File struct {
	Path        string `bun:",notnull"` // Path of the file in storage.
	ContentType string `bun:",notnull"` // MIME content type of the file.
	FileSize    int    `bun:",notnull"` // File size in bytes
}

// Thumbnail refers to a small image thumbnail derived from a larger image, video, or audio file.
type Thumbnail struct {
	Path        string `bun:",notnull"`  // Path of the file in storage.
	ContentType string `bun:",notnull"`  // MIME content type of the file.
	FileSize    int    `bun:",notnull"`  // File size in bytes
	URL         string `bun:",nullzero"` // What is the URL of the thumbnail on the local server
	RemoteURL   string `bun:",nullzero"` // What is the remote URL of the thumbnail (empty for local media)
}

// ProcessingStatus refers to how far along in the processing stage the attachment is.
type ProcessingStatus int

// MediaAttachment processing states.
const (
	ProcessingStatusReceived   ProcessingStatus = 0   // ProcessingStatusReceived indicates the attachment has been received and is awaiting processing. No thumbnail available yet.
	ProcessingStatusProcessing ProcessingStatus = 1   // ProcessingStatusProcessing indicates the attachment is currently being processed. Thumbnail is available but full media is not.
	ProcessingStatusProcessed  ProcessingStatus = 2   // ProcessingStatusProcessed indicates the attachment has been fully processed and is ready to be served.
	ProcessingStatusError      ProcessingStatus = 666 // ProcessingStatusError indicates something went wrong processing the attachment and it won't be tried again--these can be deleted.
)

// FileType refers to the file
// type of the media attaachment.
type FileType int

const (
	// MediaAttachment file types.
	FileTypeUnknown FileType = 0 // FileTypeUnknown is for unknown file types (surprise surprise!)
	FileTypeImage   FileType = 1 // FileTypeImage is for jpegs, pngs, and standard gifs
	FileTypeAudio   FileType = 2 // FileTypeAudio is for audio-only files (no video)
	FileTypeVideo   FileType = 3 // FileTypeVideo is for files with audio + visual
	FileTypeGifv    FileType = 4 // FileTypeGifv is for short video-only files (20s or less, mp4, no audio).
)

// String returns a stringified, frontend API compatible form of FileType.
func (t FileType) String() string {
	switch t {
	case FileTypeUnknown:
		return "unknown"
	case FileTypeImage:
		return "image"
	case FileTypeAudio:
		return "audio"
	case FileTypeVideo:
		return "video"
	case FileTypeGifv:
		return "gifv"
	default:
		panic("invalid filetype")
	}
}

// FileMeta describes metadata about the actual contents of the file.
type FileMeta struct {
	Original Original `bun:"embed:original_"`
	Small    Small    `bun:"embed:small_"`
	Focus    Focus    `bun:"embed:focus_"`
}

// Small can be used for a thumbnail of any media type
type Small struct {
	Width  int     // width in pixels
	Height int     // height in pixels
	Size   int     // size in pixels (width * height)
	Aspect float32 // aspect ratio (width / height)
}

// Original can be used for original metadata for any media type
type Original struct {
	Width     int      // width in pixels
	Height    int      // height in pixels
	Size      int      // size in pixels (width * height)
	Aspect    float32  // aspect ratio (width / height)
	Duration  *float32 // video-specific: duration of the video in seconds
	Framerate *float32 // video-specific: fps
	Bitrate   *uint64  // video-specific: bitrate
}

// Focus describes the 'center' of the image for display purposes.
// X and Y should each be between -1 and 1
type Focus struct {
	X float32
	Y float32
}
