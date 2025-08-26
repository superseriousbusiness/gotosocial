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
type InteractionType enumType

const (
	// WARNING: DO NOT CHANGE THE ORDER OF THESE,
	// as this will cause breakage of approvals!
	//
	// If you need to add new interaction types,
	// add them *to the end* of the list.

	InteractionLike     InteractionType = 0
	InteractionReply    InteractionType = 1
	InteractionAnnounce InteractionType = 2
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

// InteractionRequest represents one interaction request
// that is either accepted, rejected, or awaiting
// acceptance or rejection by the target account.
type InteractionRequest struct {

	// ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// ID of the status targeted by the interaction.
	TargetStatusID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Local status corresponding to TargetStatusID.
	// Column not stored in DB.
	TargetStatus *Status `bun:"-"`

	// ID of the account being interacted with.
	TargetAccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Account corresponding to TargetAccountID.
	// Column not stored in DB.
	TargetAccount *Account `bun:"-"`

	// ID of the account doing the interaction request.
	InteractingAccountID string `bun:"type:CHAR(26),nullzero,notnull"`

	// Account corresponding to InteractingAccountID.
	// Column not stored in DB.
	InteractingAccount *Account `bun:"-"`

	// URI of the Request, if this InteractionRequest originated from
	// a Request type (LikeRequest, ReplyRequest, AnnounceRequest, etc.).
	//
	// Not set if interaction request results from an interaction
	// (Like, Create (status), Announce, etc.) transmitted directly.
	InteractionRequestURI string `bun:",nullzero,unique"`

	// URI of the interaction itself.
	InteractionURI string `bun:",nullzero,notnull,unique"`

	// Type of interaction being requested.
	InteractionType InteractionType `bun:",notnull"`

	// Set if InteractionType = InteractionLike.
	// Column not stored in DB.
	Like *StatusFave `bun:"-"`

	// Set if InteractionType = InteractionReply.
	// Column not stored in DB.
	Reply *Status `bun:"-"`

	// Set if InteractionType = InteractionAnnounce.
	// Column not stored in DB.
	Announce *Status `bun:"-"`

	// If interaction request was accepted, time at which this occurred.
	AcceptedAt time.Time `bun:"type:timestamptz,nullzero"`

	// If interaction request was rejected, time at which this occurred.
	RejectedAt time.Time `bun:"type:timestamptz,nullzero"`

	// URI of the Accept (if accepted) or Reject (if rejected).
	// Field may be empty if currently neither accepted not rejected, or if
	// acceptance/rejection was implicit (ie., not resulting from an Activity).
	ResponseURI string `bun:",nullzero,unique"`

	// URI of the Authorization object (if accepted).
	//
	// Field will only be set if the interaction has been accepted.
	AuthorizationURI string `bun:",nullzero,unique"`
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

// IsPolite returns true if this interaction request was done
// "politely" with a *Request type, or false if it was done
// "impolitely" with direct send of a like, reply, or announce.
func (ir *InteractionRequest) IsPolite() bool {
	return ir.InteractionRequestURI != ""
}

// Interaction abstractly represents
// one interaction with a status, via
// liking, replying to, or boosting it.
type Interaction interface {
	GetAccount() *Account
}
