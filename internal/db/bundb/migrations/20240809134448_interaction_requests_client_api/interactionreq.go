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

type InteractionRequest struct {
	ID                   string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	StatusID             string          `bun:"type:CHAR(26),nullzero,notnull"`
	TargetAccountID      string          `bun:"type:CHAR(26),nullzero,notnull"`
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`
	InteractionURI       string          `bun:",nullzero,notnull,unique"`
	InteractionType      InteractionType `bun:",notnull"`
	AcceptedAt           time.Time       `bun:"type:timestamptz,nullzero"`
	RejectedAt           time.Time       `bun:"type:timestamptz,nullzero"`
	URI                  string          `bun:",nullzero,unique"`
}

type InteractionType int

const (
	InteractionLike InteractionType = iota
	InteractionReply
	InteractionAnnounce
)
