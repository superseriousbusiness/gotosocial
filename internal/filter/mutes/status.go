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

package mutes

import (
	"context"
	"errors"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusMuted returns whether given target status is muted for requester in the context of timeline visibility.
func (f *Filter) StatusMuted(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (muted bool, err error) {
	details, err := f.StatusMuteDetails(ctx, requester, status)
	if err != nil {
		return false, gtserror.Newf("error getting status mute details: %w", err)
	}
	return details.Mute && !details.MuteExpired(time.Now()), nil
}

// StatusNotificationsMuted returns whether notifications are muted for requester when regarding given target status.
func (f *Filter) StatusNotificationsMuted(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (muted bool, err error) {
	details, err := f.StatusMuteDetails(ctx, requester, status)
	if err != nil {
		return false, gtserror.Newf("error getting status mute details: %w", err)
	}
	return details.Notifications && !details.NotificationExpired(time.Now()), nil
}

// StatusMuteDetails returns mute details about the given status for the given requesting account.
func (f *Filter) StatusMuteDetails(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (*cache.CachedMute, error) {

	// For requester ID use a
	// fallback 'noauth' string
	// by default for lookups.
	requesterID := noauth
	if requester != nil {
		requesterID = requester.ID
	}

	// Load mute details for this requesting account about status from cache, using load callback if needed.
	details, err := f.state.Caches.Mutes.LoadOne("RequesterID,StatusID", func() (*cache.CachedMute, error) {

		// Load the mute details for given status.
		details, err := f.getStatusMuteDetails(ctx,
			requester,
			status,
		)
		if err != nil {
			if err == cache.SentinelError {
				// Filter-out our temporary
				// race-condition error.
				return &cache.CachedMute{}, nil
			}

			return nil, err
		}

		// Convert to cache details.
		return &cache.CachedMute{
			StatusID:           status.ID,
			ThreadID:           status.ThreadID,
			RequesterID:        requesterID,
			Mute:               details.mute,
			MuteExpiry:         details.muteExpiry,
			Notifications:      details.notif,
			NotificationExpiry: details.notifExpiry,
		}, nil
	}, requesterID, status.ID)
	if err != nil {
		return nil, err
	}

	return details, err
}

// getStatusMuteDetails loads muteDetails{} for the given
// status and the thread it is a part of, including any
// relevant muted parent status authors / mentions.
func (f *Filter) getStatusMuteDetails(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (
	muteDetails,
	error,
) {
	var details muteDetails

	if requester == nil {
		// Without auth, there will be no possible
		// mute to exist. Always return as 'unmuted'.
		return details, nil
	}

	// Look for a stored mute from account against thread.
	threadMute, err := f.state.DB.GetThreadMutedByAccount(ctx,
		status.ThreadID,
		requester.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return details, gtserror.Newf("db error checking thread mute: %w", err)
	}

	// Set notif mute on thread mute.
	details.notif = (threadMute != nil)

	for next := status; ; {
		// Load the mute details for 'next' status
		// in current thread, into our details obj.
		if err = f.loadOneStatusMuteDetails(ctx,
			requester,
			next,
			&details,
		); err != nil {
			return details, err
		}

		if next.InReplyToURI == "" {
			// Reached the top
			// of the thread.
			break
		}

		if next.InReplyToID == "" {
			// Parent is not yet dereferenced.
			return details, cache.SentinelError
		}

		// Check if parent is set.
		inReplyTo := next.InReplyTo
		if inReplyTo == nil {

			// Fetch next parent in conversation.
			inReplyTo, err = f.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				next.InReplyToID,
			)
			if err != nil {
				return details, gtserror.Newf("error getting status parent %s: %w", next.InReplyToURI, err)
			}
		}

		// Set next status.
		next = inReplyTo
	}

	return details, nil
}

// loadOneStatusMuteDetails loads the mute details for
// any relevant accounts to given status to the requesting
// account into the passed muteDetails object pointer.
func (f *Filter) loadOneStatusMuteDetails(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	details *muteDetails,
) error {
	// Look for mutes against related status accounts
	// by requester (e.g. author, mention targets etc).
	userMutes, err := f.getStatusRelatedUserMutes(ctx,
		requester,
		status,
	)
	if err != nil {
		return err
	}

	for _, mute := range userMutes {
		// Set as muted!
		details.mute = true

		// Set notifications as
		// muted if flag is set.
		if *mute.Notifications {
			details.notif = true
		}

		// Check for expiry data given.
		if !mute.ExpiresAt.IsZero() {

			// Update mute details expiry time if later.
			if mute.ExpiresAt.After(details.muteExpiry) {
				details.muteExpiry = mute.ExpiresAt
			}

			// Update notif details expiry time if later.
			if mute.ExpiresAt.After(details.notifExpiry) {
				details.notifExpiry = mute.ExpiresAt
			}
		}
	}

	return nil
}

// getStatusRelatedUserMutes fetches user mutes for any
// of the possible related accounts regarding this status,
// i.e. the author and any account mentioned.
func (f *Filter) getStatusRelatedUserMutes(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
) (
	[]*gtsmodel.UserMute,
	error,
) {
	if status.AccountID == requester.ID {
		// Status is by requester, we don't take
		// into account related attached user mutes.
		return nil, nil
	}

	// Preallocate a slice of worst possible case no. user mutes.
	mutes := make([]*gtsmodel.UserMute, 0, 2+len(status.Mentions))

	// Check if status is boost.
	if status.BoostOfID != "" {
		if status.BoostOf == nil {
			var err error

			// Ensure original status is loaded on boost.
			status.BoostOf, err = f.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				status.BoostOfID,
			)
			if err != nil {
				return nil, gtserror.Newf("error getting boosted status of %s: %w", status.URI, err)
			}
		}

		// Look for mute against booster.
		mute, err := f.state.DB.GetMute(
			gtscontext.SetBarebones(ctx),
			requester.ID,
			status.AccountID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("db error getting status author mute: %w", err)
		}

		if mute != nil {
			// Append author mute to total.
			mutes = append(mutes, mute)
		}

		// From here look at details
		// for original boosted status.
		status = status.BoostOf
	}

	if !status.MentionsPopulated() {
		var err error

		// Populate status mention objects before further mention checks.
		status.Mentions, err = f.state.DB.GetMentions(ctx, status.MentionIDs)
		if err != nil {
			return nil, gtserror.Newf("error populating status %s mentions: %w", status.URI, err)
		}
	}

	// Look for mute against author.
	mute, err := f.state.DB.GetMute(
		gtscontext.SetBarebones(ctx),
		requester.ID,
		status.AccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("db error getting status author mute: %w", err)
	}

	if mute != nil {
		// Append author mute to total.
		mutes = append(mutes, mute)
	}

	for _, mention := range status.Mentions {
		// Look for mute against any target mentions.
		if mention.TargetAccountID != requester.ID {

			// Look for mute against target.
			mute, err := f.state.DB.GetMute(
				gtscontext.SetBarebones(ctx),
				requester.ID,
				mention.TargetAccountID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.Newf("db error getting mention target mute: %w", err)
			}

			if mute != nil {
				// Append target mute to total.
				mutes = append(mutes, mute)
			}
		}
	}

	return mutes, nil
}
