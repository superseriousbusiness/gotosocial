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

// Like / Reply / Announce
type InteractionType int

const (
	// WARNING: DO NOT CHANGE THE ORDER OF THESE,
	// as this will cause breakage of approvals!
	//
	// If you need to add new interaction types,
	// add them *to the end* of the list.

	InteractionLike InteractionType = iota
	InteractionReply
	InteractionAnnounce
)

// Stringifies this InteractionType in a
// manner suitable for serving via the API.
func (i InteractionType) String() string {
	switch i {
	case InteractionLike:
		const text = "favourite"
		return text
	case InteractionReply:
		const text = "reply"
		return text
	case InteractionAnnounce:
		const text = "reblog"
		return text
	default:
		panic("undefined InteractionType")
	}
}

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
	StatusID             string          `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the interaction target status.
	Status               *Status         `bun:"-"`                                                           // Not stored in DB. Status being interacted with.
	AccountID            string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that owns this accept/approval
	Account              *Account        `bun:"-"`                                                           // account corresponding to accountID
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that did the interaction that this Accept targets.
	InteractingAccount   *Account        `bun:"-"`                                                           // account corresponding to targetAccountID
	InteractionURI       string          `bun:",nullzero,notnull,unique"`                                    // URI of the interacting like, reply, or announce. Unique.
	InteractionType      InteractionType `bun:",notnull"`                                                    // One of Like, Reply, or Announce.
	Like                 *StatusFave     `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionLike.
	Reply                *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionReply.
	Announce             *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionAnnounce.
	URI                  string          `bun:",nullzero,notnull,unique"`                                    // ActivityPub URI of the Accept.
}

// InteractionRequest represents one interaction (like, reply, fave)
// that is awaiting approval / rejection by the target account.
type InteractionRequest struct {
	ID                   string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	StatusID             string          `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the interaction target status.
	Status               *Status         `bun:"-"`                                                           // Not stored in DB. Status being interacted with.
	TargetAccountID      string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account being interacted with
	TargetAccount        *Account        `bun:"-"`                                                           // Not stored in DB. Account being interacted with.
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account requesting the interaction.
	InteractingAccount   *Account        `bun:"-"`                                                           // Not stored in DB. Account corresponding to targetAccountID
	InteractionURI       string          `bun:",nullzero,notnull,unique"`                                    // URI of the interacting like, reply, or announce. Unique.
	InteractionType      InteractionType `bun:",notnull"`                                                    // One of Like, Reply, or Announce.
	Like                 *StatusFave     `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionLike.
	Reply                *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionReply.
	Announce             *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionAnnounce.
}

// InteractionRejection refers to a single Reject activity sent
// to or from this instance in response to an interaction request,
// in order to reject / deny it.
type InteractionRejection struct {
	ID                   string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	StatusID             string          `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the interaction target status.
	Status               *Status         `bun:"-"`                                                           // Not stored in DB. Status being interacted with.
	AccountID            string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that owns this reject / deny.
	Account              *Account        `bun:"-"`                                                           // account corresponding to accountID
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account that did the interaction that this Reject targets.
	InteractingAccount   *Account        `bun:"-"`                                                           // account corresponding to targetAccountID
	InteractionURI       string          `bun:",nullzero,notnull"`                                           // URI of the interacting like, reply, or announce. NOT unique, as an interaction can be rejected by more than one person.
	InteractionType      InteractionType `bun:",notnull"`                                                    // One of Like, Reply, or Announce.
	URI                  string          `bun:",nullzero,notnull,unique"`                                    // ActivityPub URI of the Reject.
}
