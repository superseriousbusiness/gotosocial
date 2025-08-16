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

package media

import (
	"context"
	"errors"
	"fmt"
	"os"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-fastpath/v2"
)

// check function conformance.
var _ action.GTSAction = ListAttachments
var _ action.GTSAction = ListEmojis

// ListAttachments lists local, remote, or all attachment paths.
func ListAttachments(ctx context.Context) error {
	list, err := setupList(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure lister gets shutdown on exit.
		if err := list.shutdown(); err != nil {
			log.Errorf(ctx, "error shutting down: %v", err)
		}
	}()

	// List attachment media paths from db.
	return list.ListAttachmentPaths(ctx)
}

// ListEmojis lists local, remote, or all emoji filepaths.
func ListEmojis(ctx context.Context) error {
	list, err := setupList(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure lister gets shutdown on exit.
		if err := list.shutdown(); err != nil {
			log.Errorf(ctx, "error shutting down: %v", err)
		}
	}()

	// List emoji media paths from db.
	return list.ListEmojiPaths(ctx)
}

type list struct {
	state      *state.State
	localOnly  bool
	remoteOnly bool
}

func (l *list) ListAttachmentPaths(ctx context.Context) error {
	// Page reused for iterative
	// attachment queries, with
	// predefined limit.
	var page paging.Page
	page.Limit = 500

	// Storage base path, used for path building.
	basePath := config.GetStorageLocalBasePath()

	for {
		// Get next page of media attachments up to max ID.
		medias, err := l.state.DB.GetAttachments(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return fmt.Errorf("failed to fetch media from database: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no media or the same group is returned, we reached end.
		if len(medias) == 0 || maxID == medias[len(medias)-1].ID {
			break
		}

		// Use last ID as the next 'maxID'.
		maxID = medias[len(medias)-1].ID
		page.Max.Value = maxID

		switch {
		case l.localOnly:
			// Only print local media paths.
			for _, media := range medias {
				if media.RemoteURL == "" {
					printMediaPaths(basePath, media)
				}
			}

		case l.remoteOnly:
			// Only print remote media paths.
			for _, media := range medias {
				if media.RemoteURL != "" {
					printMediaPaths(basePath, media)
				}
			}

		default:
			// Print all known media paths.
			for _, media := range medias {
				printMediaPaths(basePath, media)
			}
		}
	}

	return nil
}

func (l *list) ListEmojiPaths(ctx context.Context) error {
	// Page reused for iterative
	// attachment queries, with
	// predefined limit.
	var page paging.Page
	page.Limit = 500

	// Storage base path, used for path building.
	basePath := config.GetStorageLocalBasePath()

	for {
		// Get the next page of emoji media up to max ID.
		emojis, err := l.state.DB.GetEmojis(ctx, &page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return fmt.Errorf("failed to fetch emojis from database: %w", err)
		}

		// Get current max ID.
		maxID := page.Max.Value

		// If no emojis or the same group is returned, we reached end.
		if len(emojis) == 0 || maxID == emojis[len(emojis)-1].ID {
			break
		}

		// Use last ID as the next 'maxID'.
		maxID = emojis[len(emojis)-1].ID
		page.Max.Value = maxID

		switch {
		case l.localOnly:
			// Only print local emoji paths.
			for _, emoji := range emojis {
				if emoji.ImageRemoteURL == "" {
					printEmojiPaths(basePath, emoji)
				}
			}

		case l.remoteOnly:
			// Only print remote emoji paths.
			for _, emoji := range emojis {
				if emoji.ImageRemoteURL != "" {
					printEmojiPaths(basePath, emoji)
				}
			}

		default:
			// Print all known emoji paths.
			for _, emoji := range emojis {
				printEmojiPaths(basePath, emoji)
			}
		}
	}

	return nil
}

func setupList(ctx context.Context) (*list, error) {
	var (
		localOnly  = config.GetAdminMediaListLocalOnly()
		remoteOnly = config.GetAdminMediaListRemoteOnly()
		state      state.State
	)

	// Validate flags.
	if localOnly && remoteOnly {
		return nil, errors.New(
			"local-only and remote-only flags cannot be true at the same time; " +
				"choose one or the other, or set neither to list all media",
		)
	}

	// Initialize caches.
	state.Caches.Init()

	// Ensure background cache tasks are running.
	if err := state.Caches.Start(); err != nil {
		return nil, fmt.Errorf("error starting caches: %w", err)
	}

	var err error

	// Only set state DB connection.
	// Don't need Actions or Workers for this.
	state.DB, err = bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbservice: %w", err)
	}

	return &list{
		state:      &state,
		localOnly:  localOnly,
		remoteOnly: remoteOnly,
	}, nil
}

func (l *list) shutdown() error {
	err := l.state.DB.Close()
	l.state.Caches.Stop()
	return err
}

// reusable path building buffer,
// only usable here as we're not
// performing concurrent writes.
var pb fastpath.Builder

// reusable string output buffer,
// only usable here as we're not
// performing concurrent writes.
var outbuf byteutil.Buffer

func printMediaPaths(basePath string, media *gtsmodel.MediaAttachment) {
	// Append file path if present.
	if media.File.Path != "" {
		path := pb.Join(basePath, media.File.Path)
		_, _ = outbuf.WriteString(path + "\n")
	}

	// Append thumb path if present.
	if media.Thumbnail.Path != "" {
		path := pb.Join(basePath, media.Thumbnail.Path)
		_, _ = outbuf.WriteString(path + "\n")
	}

	// Only write if any
	// string was prepared.
	if outbuf.Len() > 0 {
		_, _ = os.Stdout.Write(outbuf.B)
		outbuf.Reset()
	}
}

func printEmojiPaths(basePath string, emoji *gtsmodel.Emoji) {
	// Append image path if present.
	if emoji.ImagePath != "" {
		path := pb.Join(basePath, emoji.ImagePath)
		_, _ = outbuf.WriteString(path + "\n")
	}

	// Append static path if present.
	if emoji.ImageStaticPath != "" {
		path := pb.Join(basePath, emoji.ImageStaticPath)
		_, _ = outbuf.WriteString(path + "\n")
	}

	// Only write if any
	// string was prepared.
	if outbuf.Len() > 0 {
		_, _ = os.Stdout.Write(outbuf.B)
		outbuf.Reset()
	}
}
