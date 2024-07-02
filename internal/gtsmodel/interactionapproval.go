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

// InteractionApproval refers to a single Accept activity sent
// *from this instance* in response to an interaction request,
// in order to approve it.
//
// Accepts originating from remote instances are not stored
// using this format; the URI of the remote Accept is instead
// just added to the *gtsmodel.StatusFave or *gtsmodel.Status.
type InteractionApproval struct {
	ID                   string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	AccountID            string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that owns this accept/approval
	Account              *Account        `bun:"-"`                                                           // account corresponding to accountID
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that did the interaction that this Accept targets.
	InteractingAccount   *Account        `bun:"-"`                                                           // account corresponding to targetAccountID
	InteractionURI       string          `bun:",nullzero,notnull"`                                           // URI of the target like, reply, or announce
	InteractionType      InteractionType `bun:",nullzero,notnull"`                                           // One of Like, Reply, or Announce.
	URI                  string          `bun:",nullzero,notnull,unique"`                                    // ActivityPub URI of the Accept.
}

type InteractionType string

const (
	InteractionLike     InteractionType = "Like"
	InteractionReply    InteractionType = "Reply"
	InteractionAnnounce InteractionType = "Announce"
)
