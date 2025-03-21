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

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"codeberg.org/gruf/go-storage/memory"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/media/ffmpeg"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

func main() {
	ctx := context.Background()
	ctx, cncl := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cncl()

	log.SetLevel(log.ERROR)

	if len(os.Args) != 4 {
		log.Panic(ctx, "Usage: go run ./cmd/process-media <input-file> <output-processed> <output-thumbnail>")
	}

	if err := ffmpeg.InitFfprobe(ctx, 1); err != nil {
		log.Panic(ctx, err)
	}

	if err := ffmpeg.InitFfmpeg(ctx, 1); err != nil {
		log.Panic(ctx, err)
	}

	var st storage.Driver
	st.Storage = memory.Open(10, true)

	var state state.State
	state.Storage = &st

	state.Caches.Init()

	var err error

	config.SetProtocol("http")
	config.SetHost("localhost:8080")
	config.SetStorageBackend("disk")
	config.SetStorageLocalBasePath("/tmp/gotosocial")
	config.SetDbType("sqlite")
	config.SetDbAddress(":memory:")

	state.DB, err = bundb.NewBunDBService(ctx, &state)
	if err != nil {
		log.Panic(ctx, err)
	}

	if err := state.DB.CreateInstanceAccount(ctx); err != nil {
		log.Panicf(ctx, "error creating instance account: %s", err)
	}

	if err := state.DB.CreateInstanceInstance(ctx); err != nil {
		log.Panicf(ctx, "error creating instance instance: %s", err)
	}

	if err := state.DB.CreateInstanceApplication(ctx); err != nil {
		log.Panicf(ctx, "error creating instance application: %s", err)
	}

	account, err := state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		log.Panic(ctx, err)
	}

	mgr := media.NewManager(&state)

	processing, err := mgr.CreateMedia(ctx,
		account.ID,
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return os.Open(os.Args[1])
		},
		media.AdditionalMediaInfo{},
	)
	if err != nil {
		log.Panic(ctx, err)
	}

	media, err := processing.Load(ctx)
	if err != nil {
		log.Panic(ctx, err)
	}

	outputCopyable(media)
	copyFile(ctx, &st, media.File.Path, os.Args[2])
	copyFile(ctx, &st, media.Thumbnail.Path, os.Args[3])
}

func copyFile(ctx context.Context, st *storage.Driver, key string, path string) {
	rc, err := st.GetStream(ctx, key)
	if err != nil {
		if storage.IsNotFound(err) {
			return
		}
		log.Panic(ctx, err)
	}
	defer rc.Close()

	_ = os.Remove(path)

	output, err := os.Create(path)
	if err != nil {
		log.Panic(ctx, err)
	}
	defer output.Close()

	_, err = io.Copy(output, rc)
	if err != nil {
		log.Panic(ctx, err)
	}
}

func outputCopyable(media *gtsmodel.MediaAttachment) {
	var (
		now           = time.Now()
		nowStr        = now.Format(time.RFC3339)
		mediaType     string
		fileMetaExtra string
	)

	switch media.Type {
	case gtsmodel.FileTypeImage:
		mediaType = "gtsmodel.FileTypeImage"
	case gtsmodel.FileTypeVideo:
		mediaType = "gtsmodel.FileTypeVideo"
	case gtsmodel.FileTypeGifv:
		mediaType = "gtsmodel.FileTypeGifv"
	case gtsmodel.FileTypeAudio:
		mediaType = "gtsmodel.FileTypeAudio"
	case gtsmodel.FileTypeUnknown:
		mediaType = "gtsmodel.FileTypeUnknown"
	}

	if media.FileMeta.Original.Duration != nil {
		fileMetaExtra += fmt.Sprintf("\n\t\t\tDuration:  util.Ptr[float32](%f),", *media.FileMeta.Original.Duration)
	}
	if media.FileMeta.Original.Framerate != nil {
		fileMetaExtra += fmt.Sprintf("\n\t\t\tFramerate: util.Ptr[float32](%f),", *media.FileMeta.Original.Framerate)
	}
	if media.FileMeta.Original.Bitrate != nil {
		fileMetaExtra += fmt.Sprintf("\n\t\t\tBitrate:   util.Ptr[uint64](%d),", *media.FileMeta.Original.Bitrate)
	}

	fmt.Printf(`{
	ID:        "%s",
	StatusID:  "STATUS_ID_GOES_HERE",
	URL:       "%s",
	RemoteURL: "",
	CreatedAt: TimeMustParse("%s"),
	Type:      %s,
	FileMeta: gtsmodel.FileMeta{
		Original: gtsmodel.Original{
			Width:     %d,
			Height:    %d,
			Size:      %d,
			Aspect:    %f,%s
		},
		Small: gtsmodel.Small{
			Width:  %d,
			Height: %d,
			Size:   %d,
			Aspect: %f,
		},
		Focus: gtsmodel.Focus{
			X: 0,
			Y: 0,
		},
	},
	AccountID:         "ACCOUNT_ID_GOES_HERE",
	Description:       "DESCRIPTION_GOES_HERE",
	ScheduledStatusID: "",
	Blurhash:          "%s",
	Processing:        2,
	File: gtsmodel.File{
		Path:        "%s",
		ContentType: "%s",
		FileSize:    %d,
	},
	Thumbnail: gtsmodel.Thumbnail{
		Path:        "%s",
		ContentType: "%s",
		FileSize:    %d,
		URL:         "%s",
		RemoteURL:   "",
	},
	Avatar: util.Ptr(false),
	Header: util.Ptr(false),
	Cached: util.Ptr(true),
}`+"\n",
		media.ID,
		strings.ReplaceAll(media.URL, media.AccountID, "ACCOUNT_ID_GOES_HERE"),
		nowStr,
		mediaType,
		media.FileMeta.Original.Width,
		media.FileMeta.Original.Height,
		media.FileMeta.Original.Size,
		media.FileMeta.Original.Aspect,
		fileMetaExtra,
		media.FileMeta.Small.Width,
		media.FileMeta.Small.Height,
		media.FileMeta.Small.Size,
		media.FileMeta.Small.Aspect,
		media.Blurhash,
		strings.ReplaceAll(media.File.Path, media.AccountID, "ACCOUNT_ID_GOES_HERE"),
		media.File.ContentType,
		media.File.FileSize,
		strings.ReplaceAll(media.Thumbnail.Path, media.AccountID, "ACCOUNT_ID_GOES_HERE"),
		media.Thumbnail.ContentType,
		media.Thumbnail.FileSize,
		strings.ReplaceAll(media.Thumbnail.URL, media.AccountID, "ACCOUNT_ID_GOES_HERE"),
	)
}
