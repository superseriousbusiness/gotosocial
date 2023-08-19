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
	"fmt"
	"os"
	"path"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

type list struct {
	dbService db.DB
	state     *state.State
	maxID     string
	limit     int
	out       *bufio.Writer
}

// Get a list of attachment using a custom filter
func (l *list) GetAllAttachmentPaths(ctx context.Context, filter func(*gtsmodel.MediaAttachment) string) ([]string, error) {
	res := make([]string, 0, 100)
	for {
		attachments, err := l.dbService.GetAttachments(ctx, l.maxID, l.limit)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve media metadata from database: %w", err)
		}

		for _, a := range attachments {
			v := filter(a)
			if v != "" {
				res = append(res, v)
			}
		}

		// If we got less results than our limit, we've reached the
		// last page to retrieve and we can break the loop. If the
		// last batch happens to contain exactly the same amount of
		// items as the limit we'll end up doing one extra query.
		if len(attachments) < l.limit {
			break
		}

		// Grab the last ID from the batch and set it as the maxID
		// that'll be used in the next iteration so we don't get items
		// we've already seen.
		l.maxID = attachments[len(attachments)-1].ID
	}
	return res, nil
}

// Get a list of emojis using a custom filter
func (l *list) GetAllEmojisPaths(ctx context.Context, filter func(*gtsmodel.Emoji) string) ([]string, error) {
	res := make([]string, 0, 100)
	for {
		attachments, err := l.dbService.GetEmojis(ctx, l.maxID, l.limit)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve media metadata from database: %w", err)
		}

		for _, a := range attachments {
			v := filter(a)
			if v != "" {
				res = append(res, v)
			}
		}

		// If we got less results than our limit, we've reached the
		// last page to retrieve and we can break the loop. If the
		// last batch happens to contain exactly the same amount of
		// items as the limit we'll end up doing one extra query.
		if len(attachments) < l.limit {
			break
		}

		// Grab the last ID from the batch and set it as the maxID
		// that'll be used in the next iteration so we don't get items
		// we've already seen.
		l.maxID = attachments[len(attachments)-1].ID
	}
	return res, nil
}

func setupList(ctx context.Context) (*list, error) {
	var state state.State

	state.Caches.Init()
	state.Caches.Start()

	state.Workers.Start()

	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return nil, fmt.Errorf("error creating dbservice: %w", err)
	}
	state.DB = dbService

	return &list{
		dbService: dbService,
		state:     &state,
		limit:     200,
		maxID:     "",
		out:       bufio.NewWriter(os.Stdout),
	}, nil
}

func (l *list) shutdown() error {
	l.out.Flush()
	err := l.dbService.Close()
	l.state.Workers.Stop()
	l.state.Caches.Stop()
	return err
}

// Get local attachements
var ListLocalAttachment action.GTSAction = func(ctx context.Context) error {
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

	mediaPath := config.GetStorageLocalBasePath()
	attachments, err := list.GetAllAttachmentPaths(
		ctx,
		func(m *gtsmodel.MediaAttachment) string {
			if m.RemoteURL == "" {
				return path.Join(mediaPath, m.File.Path)
			}
			return ""
		})
	if err != nil {
		return err
	}

	for _, a := range attachments {
		_, _ = list.out.WriteString(a + "\n")
	}
	return nil
}

// Get remote attachements
var ListRemoteAttachment action.GTSAction = func(ctx context.Context) error {
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

	attachments, err := list.GetAllAttachmentPaths(
		ctx,
		func(m *gtsmodel.MediaAttachment) string {
			return m.RemoteURL
		})
	if err != nil {
		return err
	}

	for _, a := range attachments {
		_, _ = list.out.WriteString(a + "\n")
	}
	return nil
}

// Get local emojis
var ListLocalEmojis action.GTSAction = func(ctx context.Context) error {
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

	mediaPath := config.GetStorageLocalBasePath()
	emojis, err := list.GetAllEmojisPaths(
		ctx,
		func(e *gtsmodel.Emoji) string {
			if e.ImageRemoteURL == "" {
				return path.Join(mediaPath, e.ImagePath)
			}
			return ""
		})
	if err != nil {
		return err
	}

	for _, e := range emojis {
		_, _ = list.out.WriteString(e + "\n")
	}
	return nil
}

// Get remote emojis
var ListRemoteEmojis action.GTSAction = func(ctx context.Context) error {
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

	emojis, err := list.GetAllEmojisPaths(
		ctx,
		func(e *gtsmodel.Emoji) string {
			return e.ImageRemoteURL
		})
	if err != nil {
		return err
	}

	for _, e := range emojis {
		_, _ = list.out.WriteString(e + "\n")
	}
	return nil
}
