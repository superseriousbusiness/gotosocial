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

package model

import "mime/multipart"

// AccountExportStats models an account's stats
// specifically for the purpose of informing about
// export sizes at the /api/v1/exports/stats endpoint.
//
// swagger:model accountExportStats
type AccountExportStats struct {
	// TODO: String representation of media storage size attributed to this account.
	//
	// example: 500MB
	MediaStorage string `json:"media_storage"`

	// Number of accounts following this account.
	//
	// example: 50
	FollowersCount int `json:"followers_count"`

	// Number of accounts followed by this account.
	//
	// example: 50
	FollowingCount int `json:"following_count"`

	// Number of statuses created by this account.
	//
	// example: 81986
	StatusesCount int `json:"statuses_count"`

	// Number of lists created by this account.
	//
	// example: 10
	ListsCount int `json:"lists_count"`

	// Number of accounts blocked by this account.
	//
	// example: 15
	BlocksCount int `json:"blocks_count"`

	// Number of accounts muted by this account.
	//
	// example: 11
	MutesCount int `json:"mutes_count"`
}

// AttachmentRequest models media attachment creation parameters.
//
// swagger: ignore
type ImportRequest struct {
	// The CSV data to upload.
	Data *multipart.FileHeader `form:"data" binding:"required"`
	// Type of entries contained in the data file.
	//
	//	- `following` - accounts to follow.
	//	- `lists` - lists of accounts.
	//	- `blocks` - accounts to block.
	//	- `mutes` - accounts to mute.
	//	- `bookmarks` - statuses to bookmark.
	Type string `form:"type" binding:"required"`
	// Mode to use when creating entries from the data file:
	//	- `merge` to merge entries in file with existing entries.
	//	- `overwrite` to replace existing entries with entries in file.
	Mode string `form:"mode"`
}
