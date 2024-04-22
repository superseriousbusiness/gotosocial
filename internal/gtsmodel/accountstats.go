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

import "time"

// AccountStats models statistics
// for a remote or local account.
type AccountStats struct {
	AccountID           string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // AccountID of this AccountStats.
	RegeneratedAt       time.Time `bun:"type:timestamptz,nullzero"`                // Time this stats model was last regenerated (ie., created from scratch using COUNTs).
	FollowersCount      *int      `bun:",nullzero,notnull"`                        // Number of accounts following AccountID.
	FollowingCount      *int      `bun:",nullzero,notnull"`                        // Number of accounts followed by AccountID.
	FollowRequestsCount *int      `bun:",nullzero,notnull"`                        // Number of pending follow requests aimed at AccountID.
	StatusesCount       *int      `bun:",nullzero,notnull"`                        // Number of statuses created by AccountID.
	StatusesPinnedCount *int      `bun:",nullzero,notnull"`                        // Number of statuses pinned by AccountID.
	LastStatusAt        time.Time `bun:"type:timestamptz,nullzero"`                // Time of most recent status created by AccountID.
}
