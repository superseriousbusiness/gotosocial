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

package gtsmodel

import (
	"time"
)

// MediaAttachment represents a user-uploaded media attachment: an image/video/audio/gif that is
// somewhere in storage and that can be retrieved and served by the router.
type MediaAttachment struct {
	ID                string           `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                       // id of this item in the database
	CreatedAt         time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                // when was item created
	UpdatedAt         time.Time        `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                // when was item last updated
	StatusID          string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                        // ID of the status to which this is attached
	URL               string           `validate:"required_without=RemoteURL,omitempty,url" bun:",nullzero"`                           // Where can the attachment be retrieved on *this* server
	RemoteURL         string           `validate:"required_without=URL,omitempty,url" bun:",nullzero"`                                 // Where can the attachment be retrieved on a remote server (empty for local media)
	Type              FileType         `validate:"oneof=Image Gifv Audio Video Unknown" bun:",nullzero,notnull"`                       // Type of file (image/gifv/audio/video)
	FileMeta          FileMeta         `validate:"required" bun:",embed:,nullzero,notnull"`                                            // Metadata about the file
	AccountID         string           `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                                 // To which account does this attachment belong
	Account           *Account         `validate:"-" bun:"rel:belongs-to,join:account_id=id"`                                          // Account corresponding to accountID
	Description       string           `validate:"-" bun:""`                                                                           // Description of the attachment (for screenreaders)
	ScheduledStatusID string           `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                        // To which scheduled status does this attachment belong
	Blurhash          string           `validate:"required_if=Type Image,required_if=Type Gif,required_if=Type Video" bun:",nullzero"` // What is the generated blurhash of this attachment
	Processing        ProcessingStatus `validate:"oneof=0 1 2 666" bun:",notnull,default:2"`                                           // What is the processing status of this attachment
	File              File             `validate:"required" bun:",embed:file_,notnull,nullzero"`                                       // metadata for the whole file
	Thumbnail         Thumbnail        `validate:"required" bun:",embed:thumbnail_,notnull,nullzero"`                                  // small image thumbnail derived from a larger image, video, or audio file.
	Avatar            *bool            `validate:"-" bun:",nullzero,notnull,default:false"`                                            // Is this attachment being used as an avatar?
	Header            *bool            `validate:"-" bun:",nullzero,notnull,default:false"`                                            // Is this attachment being used as a header?
	Cached            *bool            `validate:"-" bun:",nullzero,notnull,default:false"`                                            // Is this attachment currently cached by our instance?
}

// File refers to the metadata for the whole file
type File struct {
	Path        string    `validate:"required,file" bun:",nullzero,notnull"`                               // Path of the file in storage.
	ContentType string    `validate:"required" bun:",nullzero,notnull"`                                    // MIME content type of the file.
	FileSize    int       `validate:"required" bun:",notnull"`                                             // File size in bytes
	UpdatedAt   time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // When was the file last updated.
}

// Thumbnail refers to a small image thumbnail derived from a larger image, video, or audio file.
type Thumbnail struct {
	Path        string    `validate:"required,file" bun:",nullzero,notnull"`                               // Path of the file in storage.
	ContentType string    `validate:"required" bun:",nullzero,notnull"`                                    // MIME content type of the file.
	FileSize    int       `validate:"required" bun:",notnull"`                                             // File size in bytes
	UpdatedAt   time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // When was the file last updated.
	URL         string    `validate:"required_without=RemoteURL,omitempty,url" bun:",nullzero"`            // What is the URL of the thumbnail on the local server
	RemoteURL   string    `validate:"required_without=URL,omitempty,url" bun:",nullzero"`                  // What is the remote URL of the thumbnail (empty for local media)
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

// FileType refers to the file type of the media attaachment.
type FileType string

// MediaAttachment file types.
const (
	FileTypeImage   FileType = "Image"   // FileTypeImage is for jpegs, pngs, and standard gifs
	FileTypeGifv    FileType = "Gifv"    // FileTypeGif is for soundless looping videos that behave like gifs
	FileTypeAudio   FileType = "Audio"   // FileTypeAudio is for audio-only files (no video)
	FileTypeVideo   FileType = "Video"   // FileTypeVideo is for files with audio + visual
	FileTypeUnknown FileType = "Unknown" // FileTypeUnknown is for unknown file types (surprise surprise!)
)

// FileMeta describes metadata about the actual contents of the file.
type FileMeta struct {
	Original Original `validate:"required" bun:"embed:original_"`
	Small    Small    `bun:"embed:small_"`
	Focus    Focus    `bun:"embed:focus_"`
}

// Small can be used for a thumbnail of any media type
type Small struct {
	Width  int     `validate:"required_with=Height Size Aspect"`  // width in pixels
	Height int     `validate:"required_with=Width Size Aspect"`   // height in pixels
	Size   int     `validate:"required_with=Width Height Aspect"` // size in pixels (width * height)
	Aspect float32 `validate:"required_with=Width Height Size"`   // aspect ratio (width / height)
}

// Original can be used for original metadata for any media type
type Original struct {
	Width     int      `validate:"required_with=Height Size Aspect"`  // width in pixels
	Height    int      `validate:"required_with=Width Size Aspect"`   // height in pixels
	Size      int      `validate:"required_with=Width Height Aspect"` // size in pixels (width * height)
	Aspect    float32  `validate:"required_with=Width Height Size"`   // aspect ratio (width / height)
	Duration  *float32 `validate:"-"`                                 // video-specific: duration of the video in seconds
	Framerate *float32 `validate:"-"`                                 // video-specific: fps
	Bitrate   *uint64  `validate:"-"`                                 // video-specific: bitrate
}

// Focus describes the 'center' of the image for display purposes.
// X and Y should each be between -1 and 1
type Focus struct {
	X float32 `validate:"omitempty,max=1,min=-1"`
	Y float32 `validate:"omitempty,max=1,min=-1"`
}
