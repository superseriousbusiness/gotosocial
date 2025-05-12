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

type PolicyValue string

const (
	PolicyValuePublic    PolicyValue = "public"
	PolicyValueFollowers PolicyValue = "followers"
	PolicyValueFollowing PolicyValue = "following"
	PolicyValueMutuals   PolicyValue = "mutuals"
	PolicyValueMentioned PolicyValue = "mentioned"
	PolicyValueAuthor    PolicyValue = "author"
)

type PolicyValues []PolicyValue

type InteractionPolicy struct {
	CanLike     PolicyRules
	CanReply    PolicyRules
	CanAnnounce PolicyRules
}

type PolicyRules struct {
	Always       PolicyValues
	WithApproval PolicyValues
}

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

var defaultPolicyPublic = &InteractionPolicy{
	CanLike: PolicyRules{
		Always: PolicyValues{
			PolicyValuePublic,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanReply: PolicyRules{
		Always: PolicyValues{
			PolicyValuePublic,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanAnnounce: PolicyRules{
		Always: PolicyValues{
			PolicyValuePublic,
		},
		WithApproval: make(PolicyValues, 0),
	},
}

func DefaultInteractionPolicyPublic() *InteractionPolicy {
	return defaultPolicyPublic
}

func DefaultInteractionPolicyUnlocked() *InteractionPolicy {
	// Same as public (for now).
	return defaultPolicyPublic
}

var defaultPolicyFollowersOnly = &InteractionPolicy{
	CanLike: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
			PolicyValueFollowers,
			PolicyValueMentioned,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanReply: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
			PolicyValueFollowers,
			PolicyValueMentioned,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanAnnounce: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
		},
		WithApproval: make(PolicyValues, 0),
	},
}

func DefaultInteractionPolicyFollowersOnly() *InteractionPolicy {
	return defaultPolicyFollowersOnly
}

var defaultPolicyDirect = &InteractionPolicy{
	CanLike: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
			PolicyValueMentioned,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanReply: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
			PolicyValueMentioned,
		},
		WithApproval: make(PolicyValues, 0),
	},
	CanAnnounce: PolicyRules{
		Always: PolicyValues{
			PolicyValueAuthor,
		},
		WithApproval: make(PolicyValues, 0),
	},
}

func DefaultInteractionPolicyDirect() *InteractionPolicy {
	return defaultPolicyDirect
}
