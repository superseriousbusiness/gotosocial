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

// A policy URI is GoToSocial's internal representation of
// one ActivityPub URI for an Actor or a Collection of Actors,
// specific to the domain of enforcing interaction policies.
//
// A PolicyValue can be stored in the database either as one
// of the Value constants defined below (to save space), OR as
// a full-fledged ActivityPub URI.
//
// A PolicyValue should be translated to the canonical string
// value of the represented URI when federating an item, or
// from the canonical string value of the URI when receiving
// or retrieving an item.
//
// For example, if the PolicyValue `followers` was being
// federated outwards in an interaction policy attached to an
// item created by the actor `https://example.org/users/someone`,
// then it should be translated to their followers URI when sent,
// eg., `https://example.org/users/someone/followers`.
//
// Likewise, if GoToSocial receives an item with an interaction
// policy containing `https://example.org/users/someone/followers`,
// and the item was created by `https://example.org/users/someone`,
// then the followers URI would be converted to `followers`
// for internal storage.
type PolicyValue string

const (
	// Stand-in for ActivityPub magic public URI,
	// which encompasses every possible Actor URI.
	PolicyValuePublic PolicyValue = "public"
	// Stand-in for the Followers Collection of
	// the item owner's Actor.
	PolicyValueFollowers PolicyValue = "followers"
	// Stand-in for the Following Collection of
	// the item owner's Actor.
	PolicyValueFollowing PolicyValue = "following"
	// Stand-in for the Mutuals Collection of
	// the item owner's Actor.
	//
	// (TODO: Reserved, currently unused).
	PolicyValueMutuals PolicyValue = "mutuals"
	// Stand-in for Actor URIs tagged in the item.
	PolicyValueMentioned PolicyValue = "mentioned"
	// Stand-in for the Actor URI of the item owner.
	PolicyValueAuthor PolicyValue = "author"
)

type PolicyValues []PolicyValue

// An InteractionPolicy determines which
// interactions will be accepted for an
// item, and according to what rules.
type InteractionPolicy struct {
	// Conditions in which a Like
	// interaction will be accepted
	// for an item with this policy.
	CanLike PolicyRules
	// Conditions in which a Reply
	// interaction will be accepted
	// for an item with this policy.
	CanReply PolicyRules
	// Conditions in which an Announce
	// interaction will be accepted
	// for an item with this policy.
	CanAnnounce PolicyRules
}

// PolicyRules represents the rules according
// to which a certain interaction is permitted
// to various Actor and Actor Collection URIs.
type PolicyRules struct {
	// Always is for PolicyValues who are
	// permitted to do an interaction
	// without requiring approval.
	Always PolicyValues
	// WithApproval is for PolicyValues who
	// are conditionally permitted to do
	// an interaction, pending approval.
	WithApproval PolicyValues
}
