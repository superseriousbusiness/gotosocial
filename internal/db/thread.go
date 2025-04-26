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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Thread contains functions for getting/creating
// status threads and thread mutes in the database.
type Thread interface {
	// PutThread inserts a new thread.
	PutThread(ctx context.Context, thread *gtsmodel.Thread) error

	// GetThreadMute gets a single threadMute by its ID.
	GetThreadMute(ctx context.Context, id string) (*gtsmodel.ThreadMute, error)

	// GetThreadMutedByAccount gets a threadMute targeting the
	// given thread, created by the given accountID, if it exists.
	GetThreadMutedByAccount(ctx context.Context, threadID string, accountID string) (*gtsmodel.ThreadMute, error)

	// IsThreadMutedByAccount returns true if threadID is muted
	// by given account. Empty thread ID will return false early.
	IsThreadMutedByAccount(ctx context.Context, threadID string, accountID string) (bool, error)

	// PutThreadMute inserts a new threadMute.
	PutThreadMute(ctx context.Context, threadMute *gtsmodel.ThreadMute) error

	// DeleteThreadMute deletes threadMute with the given ID.
	DeleteThreadMute(ctx context.Context, id string) error
}
