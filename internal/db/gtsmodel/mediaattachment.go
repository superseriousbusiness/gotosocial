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

package gtsmodel

import (
	"time"
)

// MediaAttachment represents a user-uploaded media attachment: an image/video/audio/gif that is
// somewhere in storage and that can be retrieved and served by the router.
type MediaAttachment struct {
	// ID of the attachment in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull,unique"`
	// ID of the status to which this is attached
	StatusID string
	// Where can the attachment be retrieved on *this* server
	URL string
	// Where can the attachment be retrieved on a remote server (empty for local media)
	RemoteURL string
	// When was the attachment created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was the attachment last updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Type of file (image/gif/audio/video)
	Type FileType `pg:",notnull"`
	// Metadata about the file
	FileMeta FileMeta
	// To which account does this attachment belong
	AccountID string `pg:",notnull"`
	// Description of the attachment (for screenreaders)
	Description string
	// To which scheduled status does this attachment belong
	ScheduledStatusID string
	// What is the generated blurhash of this attachment
	Blurhash string
	// What is the processing status of this attachment
	Processing ProcessingStatus
	// metadata for the whole file
	File File
	// small image thumbnail derived from a larger image, video, or audio file.
	Thumbnail Thumbnail
	// Is this attachment being used as an avatar?
	Avatar bool
	// Is this attachment being used as a header?
	Header bool
}

// File refers to the metadata for the whole file
type File struct {
	// What is the path of the file in storage.
	Path string
	// What is the MIME content type of the file.
	ContentType string
	// What is the size of the file in bytes.
	FileSize int
	// When was the file last updated.
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
}

// Thumbnail refers to a small image thumbnail derived from a larger image, video, or audio file.
type Thumbnail struct {
	// What is the path of the file in storage
	Path string
	// What is the MIME content type of the file.
	ContentType string
	// What is the size of the file in bytes
	FileSize int
	// When was the file last updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// What is the URL of the thumbnail on the local server
	URL string
	// What is the remote URL of the thumbnail (empty for local media)
	RemoteURL string
}

// ProcessingStatus refers to how far along in the processing stage the attachment is.
type ProcessingStatus int

const (
	// ProcessingStatusReceived: the attachment has been received and is awaiting processing. No thumbnail available yet.
	ProcessingStatusReceived ProcessingStatus = 0
	// ProcessingStatusProcessing: the attachment is currently being processed. Thumbnail is available but full media is not.
	ProcessingStatusProcessing ProcessingStatus = 1
	// ProcessingStatusProcessed: the attachment has been fully processed and is ready to be served.
	ProcessingStatusProcessed ProcessingStatus = 2
	// ProcessingStatusError: something went wrong processing the attachment and it won't be tried again--these can be deleted.
	ProcessingStatusError ProcessingStatus = 666
)

// FileType refers to the file type of the media attaachment.
type FileType string

const (
	// FileTypeImage is for jpegs and pngs
	FileTypeImage FileType = "image"
	// FileTypeGif is for native gifs and soundless videos that have been converted to gifs
	FileTypeGif FileType = "gifv"
	// FileTypeAudio is for audio-only files (no video)
	FileTypeAudio FileType = "audio"
	// FileTypeVideo is for files with audio + visual
	FileTypeVideo FileType = "video"
	// FileTypeUnknown is for unknown file types (surprise surprise!)
	FileTypeUnknown FileType = "unknown"
)

// FileMeta describes metadata about the actual contents of the file.
type FileMeta struct {
	Original Original
	Small    Small
	Focus    Focus
}

// Small can be used for a thumbnail of any media type
type Small struct {
	Width  int
	Height int
	Size   int
	Aspect float64
}

// Original can be used for original metadata for any media type
type Original struct {
	Width  int
	Height int
	Size   int
	Aspect float64
}

type Focus struct {
	X float32
	Y float32
}
