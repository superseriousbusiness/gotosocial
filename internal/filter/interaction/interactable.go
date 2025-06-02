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
	"context"
	"fmt"
	"slices"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

type matchType int

const (
	none     matchType = 0
	implicit matchType = 1
	explicit matchType = 2
)

// startedThread returns true if requester started
// the thread that the given status is part of.
// Ie., requester created the first post in the thread.
func (f *Filter) startedThread(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (bool, error) {
	parents, err := f.state.DB.GetStatusParents(ctx, status)
	if err != nil {
		return false, fmt.Errorf("db error getting parents of %s: %w", status.ID, err)
	}

	if len(parents) == 0 {
		// No parents available. Just check
		// if this status belongs to requester.
		return status.AccountID == requester.ID, nil
	}

	// Check if OG status owned by requester.
	return parents[0].AccountID == requester.ID, nil
}

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
) (*gtsmodel.PolicyCheckResult, error) {
	if requester.ID == status.AccountID {
		// Status author themself can
		// always like their own status,
		// no need for further checks.
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: util.Ptr(gtsmodel.PolicyValueAuthor),
		}, nil
	}

	switch {
	// If status has canLike sub-policy set, check against that.
	case status.InteractionPolicy != nil && status.InteractionPolicy.CanLike != nil:
		return f.checkPolicy(
			ctx,
			requester,
			status,
			status.InteractionPolicy.CanLike,
		)

	// If status is local and has no policy set,
	// check against the default policy for this
	// visibility, as we're canLike sub-policy aware.
	case *status.Local:
		return f.checkPolicy(
			ctx,
			requester,
			status,
			gtsmodel.DefaultCanLikeFor(status.Visibility),
		)

	// Otherwise, assume the status is from an
	// instance that does not use / does not care
	// about canLike sub-policy, and just return OK.
	default:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionAutomaticApproval,
		}, nil
	}
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
) (*gtsmodel.PolicyCheckResult, error) {
	if util.PtrOrValue(status.PendingApproval, false) {
		// Target status is pending approval,
		// check who started this thread.
		startedThread, err := f.startedThread(
			ctx,
			requester,
			status,
		)
		if err != nil {
			err := gtserror.Newf("error checking thread ownership: %w", err)
			return nil, err
		}

		if !startedThread {
			// If status is itself still pending approval,
			// and the requester didn't start this thread,
			// then buddy, any status that tries to reply
			// to it must be pending approval too. We do
			// this to prevent someone replying to a status
			// with a policy set that causes that reply to
			// require approval, *THEN* replying to their
			// own reply (which may not have a policy set)
			// and having the reply-to-their-own-reply go
			// through as Permitted. None of that!
			return &gtsmodel.PolicyCheckResult{
				Permission: gtsmodel.PolicyPermissionManualApproval,
			}, nil
		}
	}

	if requester.ID == status.AccountID {
		// Status author themself can
		// always reply to their own status,
		// no need for further checks.
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: util.Ptr(gtsmodel.PolicyValueAuthor),
		}, nil
	}

	// If requester is replied to by this status,
	// then just return OK, it's functionally equivalent
	// to them being mentioned, and easier to check!
	if status.InReplyToAccountID == requester.ID {
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: util.Ptr(gtsmodel.PolicyValueMentioned),
		}, nil
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
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: util.Ptr(gtsmodel.PolicyValueMentioned),
		}, nil
	}

	switch {
	// If status has canReply sub-policy set, check against that.
	case status.InteractionPolicy != nil && status.InteractionPolicy.CanReply != nil:
		return f.checkPolicy(
			ctx,
			requester,
			status,
			status.InteractionPolicy.CanReply,
		)

	// If status is local and has no policy set,
	// check against the default canReply for this
	// visibility, as we're canReply sub-policy aware.
	case *status.Local:
		return f.checkPolicy(
			ctx,
			requester,
			status,
			gtsmodel.DefaultCanReplyFor(status.Visibility),
		)

	// Otherwise, assume the status is from an
	// instance that does not use / does not care
	// about canReply sub-policy, and just return OK.
	default:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionAutomaticApproval,
		}, nil
	}
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
) (*gtsmodel.PolicyCheckResult, error) {
	if status.Visibility == gtsmodel.VisibilityDirect {
		log.Trace(ctx, "direct statuses are not boostable")
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionForbidden,
		}, nil
	}

	if requester.ID == status.AccountID {
		// Status author themself can
		// always boost non-directs,
		// no need for further checks.
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: util.Ptr(gtsmodel.PolicyValueAuthor),
		}, nil
	}

	switch {
	// If status has canAnnounce sub-policy set, check against that.
	case status.InteractionPolicy != nil && status.InteractionPolicy.CanAnnounce != nil:
		return f.checkPolicy(
			ctx,
			requester,
			status,
			status.InteractionPolicy.CanAnnounce,
		)

	// If status has no policy set but it's local,
	// check against the default policy for this
	// visibility, as we're interaction-policy aware.
	case *status.Local:
		policy := gtsmodel.DefaultInteractionPolicyFor(status.Visibility)
		return f.checkPolicy(
			ctx,
			requester,
			status,
			policy.CanAnnounce,
		)

	// Status is from an instance that does not use
	// or does not care about canAnnounce sub-policy.
	// We can boost it if it's unlisted or public.
	case status.Visibility == gtsmodel.VisibilityPublic ||
		status.Visibility == gtsmodel.VisibilityUnlocked:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionAutomaticApproval,
		}, nil

	// Not permitted by any of the
	// above checks, so it's forbidden.
	default:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionForbidden,
		}, nil
	}
}

func (f *Filter) checkPolicy(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	rules *gtsmodel.PolicyRules,
) (*gtsmodel.PolicyCheckResult, error) {

	// Wrap context to be able to
	// cache some database calls.
	fctx := new(filterctx)
	fctx.Context = ctx

	// Check if requester matches a PolicyValue
	// to be automatically approved for this.
	matchAutomatic, matchAutomaticValue, err := f.matchPolicy(fctx,
		requester,
		status,
		rules.AutomaticApproval,
	)
	if err != nil {
		return nil, gtserror.Newf("error checking policy match: %w", err)
	}

	// Check if requester matches a PolicyValue
	// to be manually approved for this.
	matchManual, _, err := f.matchPolicy(fctx,
		requester,
		status,
		rules.ManualApproval,
	)
	if err != nil {
		return nil, gtserror.Newf("error checking policy match: %w", err)
	}

	switch {

	// Prefer explicit match,
	// prioritizing automatic.
	case matchAutomatic == explicit:
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: &matchAutomaticValue,
		}, nil

	case matchManual == explicit:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionManualApproval,
		}, nil

	// Then try implicit match,
	// prioritizing automatic.
	case matchAutomatic == implicit:
		return &gtsmodel.PolicyCheckResult{
			Permission:          gtsmodel.PolicyPermissionAutomaticApproval,
			PermissionMatchedOn: &matchAutomaticValue,
		}, nil

	case matchManual == implicit:
		return &gtsmodel.PolicyCheckResult{
			Permission: gtsmodel.PolicyPermissionManualApproval,
		}, nil
	}

	// No match.
	return &gtsmodel.PolicyCheckResult{
		Permission: gtsmodel.PolicyPermissionForbidden,
	}, nil
}

// matchPolicy returns whether requesting account
// matches any of the policy values for given status,
// returning the policy it matches on and match type.
// uses a *filterctx to cache certain db results.
func (f *Filter) matchPolicy(
	ctx *filterctx,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	policyValues []gtsmodel.PolicyValue,
) (
	matchType,
	gtsmodel.PolicyValue,
	error,
) {
	var (
		match = none
		value gtsmodel.PolicyValue
	)

	for _, p := range policyValues {
		switch p {

		// Check if anyone
		// can do this.
		case gtsmodel.PolicyValuePublic:
			match = implicit
			value = gtsmodel.PolicyValuePublic

		// Check if follower
		// of status owner.
		case gtsmodel.PolicyValueFollowers:
			inFollowers, err := f.inFollowers(ctx,
				requester,
				status,
			)
			if err != nil {
				return 0, "", err
			}
			if inFollowers {
				match = implicit
				value = gtsmodel.PolicyValueFollowers
			}

		// Check if followed
		// by status owner.
		case gtsmodel.PolicyValueFollowing:
			inFollowing, err := f.inFollowing(ctx,
				requester,
				status,
			)
			if err != nil {
				return 0, "", err
			}
			if inFollowing {
				match = implicit
				value = gtsmodel.PolicyValueFollowing
			}

		// Check if replied-to by or
		// mentioned in the status.
		case gtsmodel.PolicyValueMentioned:
			if (status.InReplyToAccountID == requester.ID) ||
				status.MentionsAccount(requester.ID) {
				// Return early as we've
				// found an explicit match.
				match = explicit
				value = gtsmodel.PolicyValueMentioned
				return match, value, nil
			}

		// Check if PolicyValue specifies
		// requester explicitly.
		default:
			if string(p) == requester.URI {
				// Return early as we've
				// found an explicit match.
				match = explicit
				value = gtsmodel.PolicyValue(requester.URI)
				return match, value, nil
			}
		}
	}

	// Return either "" or "implicit",
	// and the policy value matched
	// against (if set).
	return match, value, nil
}

// inFollowers returns whether requesting account is following
// status author, uses *filterctx type for db result caching.
func (f *Filter) inFollowers(
	ctx *filterctx,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (
	bool,
	error,
) {
	if ctx.inFollowersOnce == 0 {
		var err error

		// Load the 'inFollowers' result from database.
		ctx.inFollowers, err = f.state.DB.IsFollowing(ctx,
			requester.ID,
			status.AccountID,
		)
		if err != nil {
			return false, gtserror.Newf("error checking follow status: %w", err)
		}

		// Mark value as stored.
		ctx.inFollowersOnce = 1
	}

	// Return stored value.
	return ctx.inFollowers, nil
}

// inFollowing returns whether status author is following
// requesting account, uses *filterctx for db result caching.
func (f *Filter) inFollowing(
	ctx *filterctx,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (
	bool,
	error,
) {
	if ctx.inFollowingOnce == 0 {
		var err error

		// Load the 'inFollowers' result from database.
		ctx.inFollowing, err = f.state.DB.IsFollowing(ctx,
			status.AccountID,
			requester.ID,
		)
		if err != nil {
			return false, gtserror.Newf("error checking follow status: %w", err)
		}

		// Mark value as stored.
		ctx.inFollowingOnce = 1
	}

	// Return stored value.
	return ctx.inFollowing, nil
}

// filterctx wraps a context.Context to also
// store loadable data relevant to a fillter
// operation from the database, such that it
// only needs to be loaded once IF required.
type filterctx struct {
	context.Context

	inFollowers     bool
	inFollowersOnce int32

	inFollowing     bool
	inFollowingOnce int32
}
