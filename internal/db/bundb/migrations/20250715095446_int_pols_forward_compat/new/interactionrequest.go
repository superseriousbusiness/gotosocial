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

import (
	"time"

	"github.com/uptrace/bun"
)

type InteractionRequest struct {
	// Used only for migration.
	bun.BaseModel `bun:"table:new_interaction_requests"`

	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// Removed in new model.
	// CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Renamed from "StatusID" to "TargetStatusID" in new model.
	TargetStatusID string `bun:"type:CHAR(26),nullzero,notnull"`

	TargetAccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	InteractingAccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Added in new model.
	InteractionRequestURI string `bun:",nullzero,notnull,unique"`

	InteractionURI string `bun:",nullzero,notnull,unique"`

	// Changed type from int to int16 in new model.
	InteractionType int16 `bun:",notnull"`

	// Added in new model.
	Polite *bool `bun:",nullzero,notnull,default:false"`

	AcceptedAt time.Time `bun:"type:timestamptz,nullzero"`

	RejectedAt time.Time `bun:"type:timestamptz,nullzero"`

	// Renamed from "URI" to "ResponseURI" in new model.
	ResponseURI string `bun:",nullzero,unique"`

	// Added in new model.
	AuthorizationURI string `bun:",nullzero,unique"`
}

const (
	LikeRequestSuffix     = "#LikeRequest"
	ReplyRequestSuffix    = "#ReplyRequest"
	AnnounceRequestSuffix = "#AnnounceRequest"
)
