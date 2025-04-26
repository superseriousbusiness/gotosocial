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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/media/ffmpeg"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-storage/memory"
)

func main() {
	ctx := context.Background()
	ctx, cncl := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cncl()

	log.SetLevel(log.INFO)

	if len(os.Args) != 3 {
		log.Panic(ctx, "Usage: go run ./cmd/process-emoji <input-file> <output-static>")
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

	mgr := media.NewManager(&state)

	processing, err := mgr.CreateEmoji(ctx,
		"emoji",
		"example.com",
		func(ctx context.Context) (reader io.ReadCloser, err error) {
			return os.Open(os.Args[1])
		},
		media.AdditionalEmojiInfo{
			URI: util.Ptr("example.com/emoji"),
		},
	)
	if err != nil {
		log.Panic(ctx, err)
	}

	emoji, err := processing.Load(ctx)
	if err != nil {
		log.Panic(ctx, err)
	}

	copyFile(ctx, &st, emoji.ImageStaticPath, os.Args[2])
}

func copyFile(ctx context.Context, st *storage.Driver, key string, path string) {
	rc, err := st.GetStream(ctx, key)
	if err != nil {
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
