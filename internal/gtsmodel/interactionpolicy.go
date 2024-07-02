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
// A PolicyURI can be stored in the database either as one
// of the URN constants defined below (to save space), OR as
// a full-fledged ActivityPub URI.
//
// A PolicyURI should be translated to the canonical string
// value of the represented URI when federating an item, or
// from the canonical string value of the URI when receiving
// or retrieving an item.
//
// For example, if the PolicyURI `urn:gts:followers` was being
// federated outwards in an interaction policy attached to an
// item created by the actor `https://example.org/users/someone`,
// then it should be translated to their followers URI when sent,
// eg., `https://example.org/users/someone/followers`.
//
// Likewise, if GoToSocial receives an item with an interaction
// policy containing `https://example.org/users/someone/followers`,
// and the item was created by `https://example.org/users/someone`,
// then the followers URI would be converted to `urn:gts:followers`
// for internal storage.
type PolicyURI string

const (
	// Stand-in for ActivityPub magic public URI,
	// which encompasses every possible Actor URI.
	PolicyURIPublic PolicyURI = "urn:gts:public"
	// Stand-in for the Followers Collection of
	// the item owner's Actor.
	PolicyURIFollowers PolicyURI = "urn:gts:followers"
	// Stand-in for the Following Collection of
	// the item owner's Actor.
	PolicyURIFollowing PolicyURI = "urn:gts:following"
	// Stand-in for the Mutuals Collection of
	// the item owner's Actor.
	//
	// (TODO: Reserved, currently unused).
	PolicyURIMutuals PolicyURI = "urn:gts:mutuals"
	// Stand-in for Actor URIs tagged in the item.
	PolicyURIMentioned PolicyURI = "urn:gts:mentioned"
	// Stand-in for the Actor URI of the item owner.
	PolicyURISelf PolicyURI = "urn:gts:self"
)

// FeasibleForVisibility returns true if the PolicyURI could feasibly
// be set in a policy for an item with the given visibility, otherwise
// returns false.
//
// For example, PolicyURIPublic could not be set in a policy for an
// item with visibility FollowersOnly, but could be set in a policy
// for an item with visibility Public or Unlocked.
//
// This is not prescriptive, and should be used only to guide policy
// choices. Eg., if a remote instance wants to do something wacky like
// set "anyone can interact with this status" for a Direct visibility
// status, that's their business; our normal visibility filtering will
// prevent users on our instance from actually being able to interact
// unless they can see the status anyway.
func (p PolicyURI) FeasibleForVisibility(v Visibility) bool {
	switch p {

	// Mentioned and self URNs are
	// feasible for any visibility.
	case PolicyURISelf,
		PolicyURIMentioned:
		return true

	// Followers/following/mutual URNs
	// are only feasible for items with
	// followers visibility and higher.
	case PolicyURIFollowers,
		PolicyURIFollowing:
		return v == VisibilityFollowersOnly ||
			v == VisibilityPublic ||
			v == VisibilityUnlocked

	// Public policy URN only feasible
	// for items that are To or CC public.
	case PolicyURIPublic:
		return v == VisibilityUnlocked ||
			v == VisibilityPublic

	// Any other combo
	// is probably fine.
	default:
		return true
	}
}

type PolicyURIs []PolicyURI

// PolicyResult represents the result of
// checking an Actor URI and interaction
// type against the conditions of an
// InteractionPolicy to determine if that
// interaction is permitted.
type PolicyResult int

const (
	// Interaction is forbidden for this
	// PolicyURI + interaction combination.
	PolicyResultForbidden PolicyResult = iota
	// Interaction is conditionally permitted
	// for this PolicyURI + interaction combo,
	// pending approval by the item owner.
	PolicyResultWithApproval
	// Interaction is permitted for this
	// PolicyURI + interaction combination.
	PolicyResultPermitted
)

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
	// Always is for PolicyURIs who are
	// permitted to do an interaction
	// without requiring approval.
	Always PolicyURIs
	// WithApproval is for PolicyURIs who
	// are conditionally permitted to do
	// an interaction, pending approval.
	WithApproval PolicyURIs
}

// Returns the default interaction policy
// for the given visibility level.
func DefaultInteractionPolicyFor(v Visibility) *InteractionPolicy {
	switch v {
	case VisibilityPublic:
		return DefaultInteractionPolicyPublic()
	case VisibilityUnlocked:
		return DefaultInteractionPolicyUnlocked()
	case VisibilityFollowersOnly, VisibilityMutualsOnly:
		return DefaultInteractionPolicyFollowersOnly()
	case VisibilityDirect:
		return DefaultInteractionPolicyDirect()
	default:
		panic("visibility " + v + " not recognized")
	}
}

// Returns the default interaction policy
// for a post with visibility of public.
func DefaultInteractionPolicyPublic() *InteractionPolicy {
	// Anyone can like.
	canLikeAlways := make(PolicyURIs, 1)
	canLikeAlways[0] = PolicyURIPublic

	// Unused, set empty.
	canLikeWithApproval := make(PolicyURIs, 0)

	// Anyone can reply.
	canReplyAlways := make(PolicyURIs, 1)
	canReplyAlways[0] = PolicyURIPublic

	// Unused, set empty.
	canReplyWithApproval := make(PolicyURIs, 0)

	// Anyone can announce.
	canAnnounceAlways := make(PolicyURIs, 1)
	canAnnounceAlways[0] = PolicyURIPublic

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyURIs, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of unlocked.
func DefaultInteractionPolicyUnlocked() *InteractionPolicy {
	// Same as public (for now).
	return DefaultInteractionPolicyPublic()
}

// Returns the default interaction policy for
// a post with visibility of followers only.
func DefaultInteractionPolicyFollowersOnly() *InteractionPolicy {
	// Self, followers and mentioned can like.
	canLikeAlways := make(PolicyURIs, 3)
	canLikeAlways[0] = PolicyURISelf
	canLikeAlways[1] = PolicyURIFollowers
	canLikeAlways[2] = PolicyURIMentioned

	// Unused, set empty.
	canLikeWithApproval := make(PolicyURIs, 0)

	// Self, followers and mentioned can reply.
	canReplyAlways := make(PolicyURIs, 3)
	canReplyAlways[0] = PolicyURISelf
	canReplyAlways[1] = PolicyURIFollowers
	canReplyAlways[2] = PolicyURIMentioned

	// Unused, set empty.
	canReplyWithApproval := make(PolicyURIs, 0)

	// Only self can announce.
	canAnnounceAlways := make(PolicyURIs, 1)
	canAnnounceAlways[0] = PolicyURISelf

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyURIs, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}

// Returns the default interaction policy
// for a post with visibility of direct.
func DefaultInteractionPolicyDirect() *InteractionPolicy {
	// Mentioned and self can always like.
	canLikeAlways := make(PolicyURIs, 2)
	canLikeAlways[0] = PolicyURISelf
	canLikeAlways[1] = PolicyURIMentioned

	// Unused, set empty.
	canLikeWithApproval := make(PolicyURIs, 0)

	// Mentioned and self can always reply.
	canReplyAlways := make(PolicyURIs, 2)
	canReplyAlways[0] = PolicyURISelf
	canReplyAlways[1] = PolicyURIMentioned

	// Unused, set empty.
	canReplyWithApproval := make(PolicyURIs, 0)

	// Only self can announce.
	canAnnounceAlways := make(PolicyURIs, 1)
	canAnnounceAlways[0] = PolicyURISelf

	// Unused, set empty.
	canAnnounceWithApproval := make(PolicyURIs, 0)

	return &InteractionPolicy{
		CanLike: PolicyRules{
			Always:       canLikeAlways,
			WithApproval: canLikeWithApproval,
		},
		CanReply: PolicyRules{
			Always:       canReplyAlways,
			WithApproval: canReplyWithApproval,
		},
		CanAnnounce: PolicyRules{
			Always:       canAnnounceAlways,
			WithApproval: canAnnounceWithApproval,
		},
	}
}
