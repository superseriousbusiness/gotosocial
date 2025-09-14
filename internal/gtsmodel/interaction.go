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

const (
	// Suffix to append to the URI of
	// impolite Likes to mock a LikeRequest.
	LikeRequestSuffix = "#LikeRequest"

	// Suffix to append to the URI of
	// impolite replies to mock a ReplyRequest.
	ReplyRequestSuffix = "#ReplyRequest"

	// Suffix to append to the URI of impolite
	// Announces to mock an AnnounceRequest.
	AnnounceRequestSuffix = "#AnnounceRequest"
)

// A useless function that appends two strings, this exists largely
// to indicate where a request URI is being generated as forward compatible
// with our planned polite request flow fully introduced in v0.21.0.
//
// TODO: remove this in v0.21.0. everything the linter complains about after removing this, needs updating.
func ForwardCompatibleInteractionRequestURI(interactionURI string, suffix string) string {
	return interactionURI + suffix
}

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
	// a polite Request type (LikeRequest, ReplyRequest, AnnounceRequest, etc.).
	//
	// If the interaction request originated from an interaction
	// (Like, Create (status), Announce, etc.) transmitted impolitely,
	// this will be set to a mocked URI.
	InteractionRequestURI string `bun:",nullzero,notnull,unique"`

	// URI of the interaction itself.
	InteractionURI string `bun:",nullzero,notnull,unique"`

	// Type of interaction being requested.
	InteractionType InteractionType `bun:",notnull"`

	// True if this interaction request
	// originated from a polite Request type.
	Polite *bool `bun:",nullzero,notnull,default:false"`

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
//
// The following information is not strictly needed but provides
// useful context for why interaction requests are determined to
// be either "polite" or "impolite".
//
// A "polite" interaction request flow indicates that it was
// performed with the latest currently accepted manner of doing
// things. This manner is different to that initially introduced
// by us (GoToSocial) pre-v0.20.0, as it is the result of our
// previous design going through many iterations with Mastodon
// developers as part of designing their similar quote post flow.
//
// An "impolite" interaction request flow is that produced either
// by an older interaction-policy-AWARE GoToSocial instance, or
// by any interaction-policy-UNAWARE server instance.
//
// Regarding per-version GoToSocial behaviour (summarized by tobi):
//
//   - from v0.19.0 and before, we know about and can respond only to
//     impolite interactions, and we send out impolite as well
//
//   - from v0.20.0 onwards, we know about and can respond to both
//     polite / impolite interaction requests, but we still send out impolite
//
//   - from v0.21.0 onwards, we know about and can respond to both
//     polite and impolite interaction requests, and we send out polite
func (ir *InteractionRequest) IsPolite() bool {
	return ir.Polite != nil && *ir.Polite
}

// Interaction abstractly represents
// one interaction with a status, via
// liking, replying to, or boosting it.
type Interaction interface {
	GetAccount() *Account
}
