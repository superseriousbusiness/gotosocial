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

// InteractionRequest represents one interaction (like, reply, fave)
// that is either accepted, rejected, or currently still awaiting
// acceptance or rejection by the target account.
type InteractionRequest struct {
	ID                   string          `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                    // id of this item in the database
	CreatedAt            time.Time       `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	StatusID             string          `bun:"type:CHAR(26),nullzero,notnull"`                              // ID of the interaction target status.
	Status               *Status         `bun:"-"`                                                           // Not stored in DB. Status being interacted with.
	TargetAccountID      string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account being interacted with
	TargetAccount        *Account        `bun:"-"`                                                           // Not stored in DB. Account being interacted with.
	InteractingAccountID string          `bun:"type:CHAR(26),nullzero,notnull"`                              // id of the account requesting the interaction.
	InteractingAccount   *Account        `bun:"-"`                                                           // Not stored in DB. Account corresponding to targetAccountID
	InteractionURI       string          `bun:",nullzero,notnull,unique"`                                    // URI of the interacting like, reply, or announce. Unique (only one interaction request allowed per interaction URI).
	InteractionType      InteractionType `bun:",notnull"`                                                    // One of Like, Reply, or Announce.
	Like                 *StatusFave     `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionLike.
	Reply                *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionReply.
	Announce             *Status         `bun:"-"`                                                           // Not stored in DB. Only set if InteractionType = InteractionAnnounce.
	AcceptedAt           time.Time       `bun:"type:timestamptz,nullzero"`                                   // If interaction request was accepted, time at which this occurred.
	RejectedAt           time.Time       `bun:"type:timestamptz,nullzero"`                                   // If interaction request was rejected, time at which this occurred.

	// ActivityPub URI of the Accept (if accepted) or Reject (if rejected).
	// Field may be empty if currently neither accepted not rejected, or if
	// acceptance/rejection was implicit (ie., not resulting from an Activity).
	URI string `bun:",nullzero,unique"`
}

// IsHandled returns true if interaction
// request has been neither accepted or rejected.
func (ir *InteractionRequest) IsPending() bool {
	return !ir.IsAccepted() && !ir.IsRejected()
}

// IsAccepted returns true if this
// interaction request has been accepted.
func (ir *InteractionRequest) IsAccepted() bool {
	return !ir.AcceptedAt.IsZero()
}

// IsRejected returns true if this
// interaction request has been rejected.
func (ir *InteractionRequest) IsRejected() bool {
	return !ir.RejectedAt.IsZero()
}
