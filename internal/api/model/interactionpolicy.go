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

// One interaction policy entry for a status.
//
// It can be EITHER one of the internal URNs listed below, OR a full-fledged ActivityPub URI of an Actor, like "https://example.org/users/some_user".
//
// Note that the URN entries (beginning with "urn") are specific to *this status* on *this instance* as viewed by *this account*
//
// In other words, they are for client API use only, and are not intended to be globally resolvable.
//
// Internal URNs:
//
//   - urn:mastodon:public    - Public, aka anyone who can see the status according to its visibility level.
//   - urn:mastodon:followers - Followers of the status author.
//   - urn:mastodon:following - People followed by the status author.
//   - urn:mastodon:mutuals   - Mutual follows of the status author (reserved, unused).
//   - urn:mastodon:mentioned - Accounts mentioned in, or replied-to by, the status.
//   - urn:mastodon:author    - The status author themself.
//   - urn:mastodon:me        - If request was made with an authorized user, "me" represents the user who made the request and is now looking at this interaction policy.
//
// swagger:model interactionPolicyURI
type PolicyURI string

const (
	PolicyURIPublic    PolicyURI = "urn:mastodon:public"    // Public, aka anyone who can see the status according to its visibility level.
	PolicyURIFollowers PolicyURI = "urn:mastodon:followers" // Followers of the status author.
	PolicyURIFollowing PolicyURI = "urn:mastodon:following" // People followed by the status author.
	PolicyURIMutuals   PolicyURI = "urn:mastodon:mutuals"   // Mutual follows of the status author (reserved, unused).
	PolicyURIMentioned PolicyURI = "urn:mastodon:mentioned" // Accounts mentioned in, or replied-to by, the status.
	PolicyURIAuthor    PolicyURI = "urn:mastodon:author"    // The status author themself.
	PolicyURIMe        PolicyURI = "urn:mastodon:me"        // If request was made with an authorized user, "me" represents the user who made the request and is now looking at this interaction policy.
)

// Rules for one interaction type.
//
// swagger:model interactionPolicyRules
type PolicyRules struct {
	// Policy entries for accounts that can always do this type of interaction.
	Always []PolicyURI `form:"always[]" json:"always"`
	// Policy entries for accounts that require approval to do this type of interaction.
	WithApproval []PolicyURI `form:"with_approval[]" json:"with_approval"`
}

// Interaction policy of a status.
//
// swagger:model interactionPolicy
type InteractionPolicy struct {
	// Rules for who can favourite this status.
	CanFavourite PolicyRules `form:"can_favourite" json:"can_favourite"`
	// Rules for who can reply to this status.
	CanReply PolicyRules `form:"can_reply" json:"can_reply"`
	// Rules for who can reblog this status.
	CanReblog PolicyRules `form:"can_reblog" json:"can_reblog"`
}

// Default interaction policies to use for new statuses by requesting account.
//
// swagger:model defaultPolicies
type DefaultPolicies struct {
	// TODO: Add mutuals only default.

	// Default policy for new direct visibility statuses.
	Direct InteractionPolicy `json:"direct"`
	// Default policy for new private/followers-only visibility statuses.
	Private InteractionPolicy `json:"private"`
	// Default policy for new unlisted/unlocked visibility statuses.
	Unlisted InteractionPolicy `json:"unlisted"`
	// Default policy for new public visibility statuses.
	Public InteractionPolicy `json:"public"`
}

// swagger:parameters policiesDefaultsUpdate
type UpdateInteractionPoliciesRequest struct {
	// Default policy for new direct visibility statuses.
	// Value `null` or omitted property resets policy to original default.
	//
	// in: formData
	// nullable: true
	Direct *InteractionPolicy `form:"direct" json:"direct"`
	// Default policy for new private/followers-only visibility statuses.
	// Value `null` or omitted property resets policy to original default.
	//
	// in: formData
	// nullable: true
	Private *InteractionPolicy `form:"private" json:"private"`
	// Default policy for new unlisted/unlocked visibility statuses.
	// Value `null` or omitted property resets policy to original default.
	//
	// in: formData
	// nullable: true
	Unlisted *InteractionPolicy `form:"unlisted" json:"unlisted"`
	// Default policy for new public visibility statuses.
	// Value `null` or omitted property resets policy to original default.
	//
	// in: formData
	// nullable: true
	Public *InteractionPolicy `form:"public" json:"public"`
}
