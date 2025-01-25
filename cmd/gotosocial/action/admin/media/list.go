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
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type list struct {
	dbService  db.DB
	state      *state.State
	page       paging.Page
	localOnly  bool
	remoteOnly bool
	out        *bufio.Writer
}

// Get a list of attachment using a custom filter
func (l *list) GetAllAttachmentPaths(ctx context.Context, filter func(*gtsmodel.MediaAttachment) string) ([]string, error) {
	res := make([]string, 0, 100)

	for {
		// Get the next page of media attachments up to max ID.
		attachments, err := l.dbService.GetAttachments(ctx, &l.page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, fmt.Errorf("failed to retrieve media metadata from database: %w", err)
		}

		// Get current max ID.
		maxID := l.page.Max.Value

		// If no attachments or the same group is returned, we reached the end.
		if len(attachments) == 0 || maxID == attachments[len(attachments)-1].ID {
			break
		}

		// Use last ID as the next 'maxID' value.
		maxID = attachments[len(attachments)-1].ID
		l.page.Max = paging.MaxID(maxID)

		for _, a := range attachments {
			v := filter(a)
			if v != "" {
				res = append(res, v)
			}
		}
	}
	return res, nil
}

// Get a list of emojis using a custom filter
func (l *list) GetAllEmojisPaths(ctx context.Context, filter func(*gtsmodel.Emoji) string) ([]string, error) {
	res := make([]string, 0, 100)
	for {
		// Get the next page of emoji media up to max ID.
		attachments, err := l.dbService.GetEmojis(ctx, &l.page)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, fmt.Errorf("failed to retrieve media metadata from database: %w", err)
		}

		// Get current max ID.
		maxID := l.page.Max.Value

		// If no attachments or the same group is returned, we reached the end.
		if len(attachments) == 0 || maxID == attachments[len(attachments)-1].ID {
			break
		}

		// Use last ID as the next 'maxID' value.
		maxID = attachments[len(attachments)-1].ID
		l.page.Max = paging.MaxID(maxID)

		for _, a := range attachments {
			v := filter(a)
			if v != "" {
				res = append(res, v)
			}
		}
	}
	return res, nil
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

	state.Caches.Init()
	state.Caches.Start()

	// Only set state DB connection.
	// Don't need Actions or Workers for this.
	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbservice: %w", err)
	}
	state.DB = dbService

	return &list{
		dbService:  dbService,
		state:      &state,
		page:       paging.Page{Limit: 200},
		localOnly:  localOnly,
		remoteOnly: remoteOnly,
		out:        bufio.NewWriter(os.Stdout),
	}, nil
}

func (l *list) shutdown() error {
	l.out.Flush()
	err := l.dbService.Close()
	l.state.Caches.Stop()
	return err
}

// ListAttachments lists local, remote, or all attachment paths.
var ListAttachments action.GTSAction = func(ctx context.Context) error {
	list, err := setupList(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure lister gets shutdown on exit.
		if err := list.shutdown(); err != nil {
			log.Error(ctx, err)
		}
	}()

	var (
		mediaPath = config.GetStorageLocalBasePath()
		filter    func(*gtsmodel.MediaAttachment) string
	)

	switch {
	case list.localOnly:
		filter = func(m *gtsmodel.MediaAttachment) string {
			if m.RemoteURL != "" {
				// Remote, not
				// interested.
				return ""
			}

			return path.Join(mediaPath, m.File.Path)
		}

	case list.remoteOnly:
		filter = func(m *gtsmodel.MediaAttachment) string {
			if m.RemoteURL == "" {
				// Local, not
				// interested.
				return ""
			}

			return path.Join(mediaPath, m.File.Path)
		}

	default:
		filter = func(m *gtsmodel.MediaAttachment) string {
			return path.Join(mediaPath, m.File.Path)
		}
	}

	attachments, err := list.GetAllAttachmentPaths(ctx, filter)
	if err != nil {
		return err
	}

	for _, a := range attachments {
		_, _ = list.out.WriteString(a + "\n")
	}
	return nil
}

// ListEmojis lists local, remote, or all emoji filepaths.
var ListEmojis action.GTSAction = func(ctx context.Context) error {
	list, err := setupList(ctx)
	if err != nil {
		return err
	}

	defer func() {
		// Ensure lister gets shutdown on exit.
		if err := list.shutdown(); err != nil {
			log.Error(ctx, err)
		}
	}()

	var (
		mediaPath = config.GetStorageLocalBasePath()
		filter    func(*gtsmodel.Emoji) string
	)

	switch {
	case list.localOnly:
		filter = func(e *gtsmodel.Emoji) string {
			if e.ImageRemoteURL != "" {
				// Remote, not
				// interested.
				return ""
			}

			return path.Join(mediaPath, e.ImagePath)
		}

	case list.remoteOnly:
		filter = func(e *gtsmodel.Emoji) string {
			if e.ImageRemoteURL == "" {
				// Local, not
				// interested.
				return ""
			}

			return path.Join(mediaPath, e.ImagePath)
		}

	default:
		filter = func(e *gtsmodel.Emoji) string {
			return path.Join(mediaPath, e.ImagePath)
		}
	}

	emojis, err := list.GetAllEmojisPaths(ctx, filter)
	if err != nil {
		return err
	}

	for _, e := range emojis {
		_, _ = list.out.WriteString(e + "\n")
	}
	return nil
}
