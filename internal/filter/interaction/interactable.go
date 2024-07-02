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

package interaction

import (
	"cmp"
	"context"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// StatusLikeable checks if the given status
// is likeable by the requester account.
//
// Callers to this function should have already
// checked the visibility of status to requester,
// including taking account of blocks, as this
// function does not do visibility checks, only
// interaction policy checks.
func (f *Filter) StatusLikeable(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (gtsmodel.PolicyResult, error) {
	if requester.ID == status.AccountID {
		// Status author themself can
		// always like their own status,
		// no need for further checks.
		return gtsmodel.PolicyResultPermitted, nil
	}

	// If status has policy set, check against that.
	// Otherwise, check against the default policy
	// for this visibility.
	var policy *gtsmodel.InteractionPolicy
	if status.InteractionPolicy != nil {
		policy = status.InteractionPolicy
	} else {
		policy = gtsmodel.DefaultInteractionPolicyFor(status.Visibility)
	}

	return f.checkPolicy(ctx, requester, status, policy.CanLike)
}

// StatusReplyable checks if the given status
// is replyable by the requester account.
//
// Callers to this function should have already
// checked the visibility of status to requester,
// including taking account of blocks, as this
// function does not do visibility checks, only
// interaction policy checks.
func (f *Filter) StatusReplyable(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (gtsmodel.PolicyResult, error) {
	if requester.ID == status.AccountID {
		// Status author themself can
		// always reply to their own status,
		// no need for further checks.
		return gtsmodel.PolicyResultPermitted, nil
	}

	// Check if requester mentioned by this status.
	//
	// Prefer checking by ID, fall back to URI, URL,
	// or NameString for not-yet enriched statuses.
	mentioned := slices.ContainsFunc(
		status.Mentions,
		func(m *gtsmodel.Mention) bool {
			switch {

			// Check by ID - most accurate.
			case m.TargetAccountID != "":
				return m.TargetAccountID == requester.ID

			// Check by URI - also accurate.
			case m.TargetAccountURI != "":
				return m.TargetAccountURI == requester.URI

			// Check by URL - probably accurate.
			case m.TargetAccountURL != "":
				return m.TargetAccountURL == requester.URL

			// Fall back to checking by namestring.
			case m.NameString != "":
				username, host, err := util.ExtractNamestringParts(m.NameString)
				if err != nil {
					log.Debugf(ctx, "error checking if mentioned: %v", err)
					return false
				}

				if requester.IsLocal() {
					// Local requester has empty string
					// domain so check using config.
					return username == requester.Username &&
						(host == config.GetHost() || host == config.GetAccountDomain())
				}

				// Remote requester has domain set.
				return username == requester.Username &&
					host == requester.Domain

			default:
				// Not mentioned.
				return false
			}
		},
	)

	if mentioned {
		// A mentioned account can always
		// reply, no need for further checks.
		return gtsmodel.PolicyResultPermitted, nil
	}

	// If status has policy set, check against that.
	// Otherwise, check against the default policy
	// for this visibility.
	var policy *gtsmodel.InteractionPolicy
	if status.InteractionPolicy != nil {
		policy = status.InteractionPolicy
	} else {
		policy = gtsmodel.DefaultInteractionPolicyFor(status.Visibility)
	}

	return f.checkPolicy(ctx, requester, status, policy.CanReply)
}

// StatusBoostable checks if the given status
// is boostable by the requester account.
//
// Callers to this function should have already
// checked the visibility of status to requester,
// including taking account of blocks, as this
// function does not do visibility checks, only
// interaction policy checks.
func (f *Filter) StatusBoostable(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (gtsmodel.PolicyResult, error) {
	if status.Visibility == gtsmodel.VisibilityDirect {
		log.Trace(ctx, "direct statuses are not boostable")
		return gtsmodel.PolicyResultForbidden, nil
	}

	if requester.ID == status.AccountID {
		// Status author themself can
		// always boost non-directs,
		// no need for further checks.
		return gtsmodel.PolicyResultPermitted, nil
	}

	// If status has policy set, check against that.
	// Otherwise, check against the default policy
	// for this visibility.
	var policy *gtsmodel.InteractionPolicy
	if status.InteractionPolicy != nil {
		policy = status.InteractionPolicy
	} else {
		policy = gtsmodel.DefaultInteractionPolicyFor(status.Visibility)
	}

	return f.checkPolicy(ctx, requester, status, policy.CanAnnounce)
}

func (f *Filter) checkPolicy(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	rules gtsmodel.PolicyRules,
) (gtsmodel.PolicyResult, error) {
	var (
		inFollowers    bool
		inFollowersErr error
		inFollowing    bool
		inFollowingErr error
	)

	// Save DB calls by wrapping inFollowers
	// check in DoOnce so we only call it once.
	inFollowersF := util.DoOnce(func() {
		b, err := f.state.DB.IsFollowing(ctx, requester.ID, status.AccountID)
		if err != nil {
			inFollowersErr = gtserror.Newf(
				"db error checking if %s follows %s: %w",
				requester.ID, status.AccountID, inFollowersErr,
			)
		}
		inFollowers = b
	})

	// Save DB calls by wrapping inFollowing
	// check in DoOnce so we only call it once.
	inFollowingF := util.DoOnce(func() {
		b, err := f.state.DB.IsFollowing(ctx, status.AccountID, requester.ID)
		if err != nil {
			inFollowingErr = gtserror.Newf(
				"db error checking if %s follows %s: %w",
				status.AccountID, requester.ID, inFollowingErr,
			)
		}
		inFollowing = b
	})

	const (
		no       = ""
		implicit = "implicit"
		explicit = "explicit"
	)

	matches := func(PolicyURIs []gtsmodel.PolicyURI) string {
		match := no
		for _, p := range PolicyURIs {
			switch p {

			// Check if anyone
			// can do this.
			case gtsmodel.PolicyURIPublic:
				match = implicit

			// Check if follower
			// of status owner.
			case gtsmodel.PolicyURIFollowers:
				inFollowersF()
				if inFollowers {
					match = implicit
				}

			// Check if followed
			// by status owner.
			case gtsmodel.PolicyURIFollowing:
				inFollowingF()
				if inFollowing {
					match = implicit
				}

			// Check if replied-to by or
			// mentioned in the status.
			case gtsmodel.PolicyURIMentioned:
				if (status.InReplyToAccountID == requester.ID) ||
					status.MentionsAccount(requester.ID) {
					// Return early as we've
					// found an explicit match.
					return explicit
				}

			// Check if PolicyURI specifies
			// requester explicitly.
			default:
				if string(p) == requester.URI {
					// Return early as we've
					// found an explicit match.
					return explicit
				}
			}
		}

		// Return either ""
		// or "implicit".
		return match
	}

	// Check if requester matches a PolicyURI
	// to be always allowed to do this.
	matchAlways := matches(rules.Always)

	// Check if requester matches a PolicyURI
	// to be allowed to do this pending approval.
	matchWithApproval := matches(rules.WithApproval)

	// Start preparing result.
	result := gtsmodel.PolicyResultForbidden

	// Return early if we
	// encountered an error.
	if err := cmp.Or(
		inFollowersErr,
		inFollowingErr,
	); err != nil {
		return result, err
	}

	switch {

	// Prefer explicit match,
	// prioritizing "always".
	case matchAlways == explicit:
		result = gtsmodel.PolicyResultPermitted

	case matchWithApproval == explicit:
		result = gtsmodel.PolicyResultWithApproval

	// Then try implicit match,
	// prioritizing "always".
	case matchAlways == implicit:
		result = gtsmodel.PolicyResultPermitted

	case matchWithApproval == implicit:
		result = gtsmodel.PolicyResultWithApproval
	}

	return result, nil
}
