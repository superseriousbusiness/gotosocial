/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package visibility

import (
	"context"
	"fmt"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (f *filter) StatusVisible(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error) {
	l := log.WithFields(kv.Fields{{"statusID", targetStatus.ID}}...)

	// Fetch any relevant accounts for the target status
	const getBoosted = true
	relevantAccounts, err := f.relevantAccounts(ctx, targetStatus, getBoosted)
	if err != nil {
		l.Debugf("error pulling relevant accounts for status %s: %s", targetStatus.ID, err)
		return false, fmt.Errorf("StatusVisible: error pulling relevant accounts for status %s: %s", targetStatus.ID, err)
	}

	// Check we have determined a target account
	targetAccount := relevantAccounts.Account
	if targetAccount == nil {
		l.Trace("target account is not set")
		return false, nil
	}

	// Check for domain blocks among relevant accounts
	domainBlocked, err := f.domainBlockedRelevant(ctx, relevantAccounts)
	if err != nil {
		l.Debugf("error checking domain block: %s", err)
		return false, fmt.Errorf("error checking domain block: %s", err)
	} else if domainBlocked {
		return false, nil
	}

	// if target account is suspended then don't show the status
	if !targetAccount.SuspendedAt.IsZero() {
		l.Trace("target account suspended at is not zero")
		return false, nil
	}

	// if the target user doesn't exist (anymore) then the status also shouldn't be visible
	// note: we only do this for local users
	if targetAccount.Domain == "" {
		targetUser, err := f.db.GetUserByAccountID(ctx, targetAccount.ID)
		if err != nil {
			l.Debug("target user could not be selected")
			if err == db.ErrNoEntries {
				return false, nil
			}
			return false, fmt.Errorf("StatusVisible: db error selecting user for local target account %s: %s", targetAccount.ID, err)
		}

		// if target user is disabled, not yet approved, or not confirmed then don't show the status
		// (although in the latter two cases it's unlikely they posted a status yet anyway, but you never know!)
		if *targetUser.Disabled || !*targetUser.Approved || targetUser.ConfirmedAt.IsZero() {
			l.Trace("target user is disabled, not approved, or not confirmed")
			return false, nil
		}
	}

	// If requesting account is nil, that means whoever requested the status didn't auth, or their auth failed.
	// In this case, we can still serve the status if it's public, otherwise we definitely shouldn't.
	if requestingAccount == nil {
		if targetStatus.Visibility == gtsmodel.VisibilityPublic {
			return true, nil
		}
		l.Trace("requesting account is nil but the target status isn't public")
		return false, nil
	}

	// if the requesting user doesn't exist (anymore) then the status also shouldn't be visible
	// note: we only do this for local users
	if requestingAccount.Domain == "" {
		requestingUser, err := f.db.GetUserByAccountID(ctx, requestingAccount.ID)
		if err != nil {
			// if the requesting account is local but doesn't have a corresponding user in the db this is a problem
			l.Debug("requesting user could not be selected")
			if err == db.ErrNoEntries {
				return false, nil
			}
			return false, fmt.Errorf("StatusVisible: db error selecting user for local requesting account %s: %s", requestingAccount.ID, err)
		}
		// okay, user exists, so make sure it has full privileges/is confirmed/approved
		if *requestingUser.Disabled || !*requestingUser.Approved || requestingUser.ConfirmedAt.IsZero() {
			l.Trace("requesting account is local but corresponding user is either disabled, not approved, or not confirmed")
			return false, nil
		}
	}

	// if requesting account is suspended then don't show the status -- although they probably shouldn't have gotten
	// this far (ie., been authed) in the first place: this is just for safety.
	if !requestingAccount.SuspendedAt.IsZero() {
		l.Trace("requesting account is suspended")
		return false, nil
	}

	// if the target status belongs to the requesting account, they should always be able to view it at this point
	if targetStatus.AccountID == requestingAccount.ID {
		return true, nil
	}

	// At this point we have a populated targetAccount, targetStatus, and requestingAccount, so we can check for blocks and whathaveyou
	// First check if a block exists directly between the target account (which authored the status) and the requesting account.
	if blocked, err := f.db.IsBlocked(ctx, targetAccount.ID, requestingAccount.ID, true); err != nil {
		l.Debugf("something went wrong figuring out if the accounts have a block: %s", err)
		return false, err
	} else if blocked {
		// don't allow the status to be viewed if a block exists in *either* direction between these two accounts, no creepy stalking please
		l.Trace("a block exists between requesting account and target account")
		return false, nil
	}

	// If not in reply to the requesting account, check if inReplyToAccount is blocked
	if relevantAccounts.InReplyToAccount != nil && relevantAccounts.InReplyToAccount.ID != requestingAccount.ID {
		if blocked, err := f.db.IsBlocked(ctx, relevantAccounts.InReplyToAccount.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and reply to account")
			return false, nil
		}
	}

	// status boosts accounts id
	if relevantAccounts.BoostedAccount != nil {
		if blocked, err := f.db.IsBlocked(ctx, relevantAccounts.BoostedAccount.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and boosted account")
			return false, nil
		}
	}

	// status boosts a reply to account id
	if relevantAccounts.BoostedInReplyToAccount != nil {
		if blocked, err := f.db.IsBlocked(ctx, relevantAccounts.BoostedInReplyToAccount.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and boosted reply to account")
			return false, nil
		}
	}

	// boost mentions accounts
	for _, a := range relevantAccounts.BoostedMentionedAccounts {
		if a == nil {
			continue
		}
		if blocked, err := f.db.IsBlocked(ctx, a.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and a boosted mentioned account")
			return false, nil
		}
	}

	// Iterate mentions to check for blocks or requester mentions
	isMentioned, blockAmongMentions := false, false
	for _, a := range relevantAccounts.MentionedAccounts {
		if a == nil {
			continue
		}

		if blocked, err := f.db.IsBlocked(ctx, a.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			blockAmongMentions = true
			break
		}

		if a.ID == requestingAccount.ID {
			isMentioned = true
		}
	}

	if blockAmongMentions {
		l.Trace("a block exists between requesting account and a mentioned account")
		return false, nil
	} else if isMentioned {
		// Requester mentioned, should always be visible
		return true, nil
	}

	// at this point we know neither account blocks the other, or another account mentioned or otherwise referred to in the status
	// that means it's now just a matter of checking the visibility settings of the status itself
	switch targetStatus.Visibility {
	case gtsmodel.VisibilityPublic, gtsmodel.VisibilityUnlocked:
		// no problem here
	case gtsmodel.VisibilityFollowersOnly:
		// Followers-only post, check for a one-way follow to target
		follows, err := f.db.IsFollowing(ctx, requestingAccount, targetAccount)
		if err != nil {
			return false, err
		}
		if !follows {
			l.Trace("requested status is followers only but requesting account is not a follower")
			return false, nil
		}
	case gtsmodel.VisibilityMutualsOnly:
		// Mutuals-only post, check for a mutual follow
		mutuals, err := f.db.IsMutualFollowing(ctx, requestingAccount, targetAccount)
		if err != nil {
			return false, err
		}
		if !mutuals {
			l.Trace("requested status is mutuals only but accounts aren't mufos")
			return false, nil
		}
	case gtsmodel.VisibilityDirect:
		l.Trace("requesting account requests a direct status it's not mentioned in")
		return false, nil // it's not mentioned -_-
	}

	// If we reached here, all is okay
	return true, nil
}

func (f *filter) StatusesVisible(ctx context.Context, statuses []*gtsmodel.Status, requestingAccount *gtsmodel.Account) ([]*gtsmodel.Status, error) {
	filtered := []*gtsmodel.Status{}
	for _, s := range statuses {
		visible, err := f.StatusVisible(ctx, s, requestingAccount)
		if err != nil {
			return nil, err
		}
		if visible {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}
