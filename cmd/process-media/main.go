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
	"io"
	"os"
	"os/signal"
	"syscall"

	"codeberg.org/gruf/go-storage/memory"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
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

	log.SetLevel(log.INFO)

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

	config.SetHost("example.com")
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
