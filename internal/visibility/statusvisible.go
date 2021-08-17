package visibility

import (
	"errors"

	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusVisible(targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":     "StatusVisible",
		"statusID": targetStatus.ID,
	})

	relevantAccounts, err := f.pullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		l.Debugf("error pulling relevant accounts for status %s: %s", targetStatus.ID, err)
		return false, fmt.Errorf("error pulling relevant accounts for status %s: %s", targetStatus.ID, err)
	}

	domainBlocked, err := f.domainBlockedRelevant(relevantAccounts)
	if err != nil {
		l.Debugf("error checking domain block: %s", err)
		return false, fmt.Errorf("error checking domain block: %s", err)
	}

	if domainBlocked {
		return false, nil
	}

	targetAccount := relevantAccounts.StatusAuthor
	// if target account is suspended then don't show the status
	if !targetAccount.SuspendedAt.IsZero() {
		l.Trace("target account suspended at is not zero")
		return false, nil
	}

	// if the target user doesn't exist (anymore) then the status also shouldn't be visible
	// note: we only do this for local users
	if targetAccount.Domain == "" {
		targetUser := &gtsmodel.User{}
		if err := f.db.GetWhere([]db.Where{{Key: "account_id", Value: targetAccount.ID}}, targetUser); err != nil {
			l.Debug("target user could not be selected")
			if err == db.ErrNoEntries {
				return false, nil
			}
			return false, fmt.Errorf("StatusVisible: db error selecting user for local target account %s: %s", targetAccount.ID, err)
		}

		// if target user is disabled, not yet approved, or not confirmed then don't show the status
		// (although in the latter two cases it's unlikely they posted a status yet anyway, but you never know!)
		if targetUser.Disabled || !targetUser.Approved || targetUser.ConfirmedAt.IsZero() {
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
		requestingUser := &gtsmodel.User{}
		if err := f.db.GetWhere([]db.Where{{Key: "account_id", Value: requestingAccount.ID}}, requestingUser); err != nil {
			// if the requesting account is local but doesn't have a corresponding user in the db this is a problem
			l.Debug("requesting user could not be selected")
			if err == db.ErrNoEntries {
				return false, nil
			}
			return false, fmt.Errorf("StatusVisible: db error selecting user for local requesting account %s: %s", requestingAccount.ID, err)
		}
		// okay, user exists, so make sure it has full privileges/is confirmed/approved
		if requestingUser.Disabled || !requestingUser.Approved || requestingUser.ConfirmedAt.IsZero() {
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
	if blocked, err := f.db.Blocked(targetAccount.ID, requestingAccount.ID, true); err != nil {
		l.Debugf("something went wrong figuring out if the accounts have a block: %s", err)
		return false, err
	} else if blocked {
		// don't allow the status to be viewed if a block exists in *either* direction between these two accounts, no creepy stalking please
		l.Trace("a block exists between requesting account and target account")
		return false, nil
	}

	// status replies to account id
	if relevantAccounts.ReplyToAccount != nil && relevantAccounts.ReplyToAccount.ID != requestingAccount.ID {
		if blocked, err := f.db.Blocked(relevantAccounts.ReplyToAccount.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and reply to account")
			return false, nil
		}

		// check reply to ID
		if targetStatus.InReplyToID != "" && (targetStatus.Visibility == gtsmodel.VisibilityFollowersOnly || targetStatus.Visibility == gtsmodel.VisibilityDirect) {
			followsRepliedAccount, err := f.db.Follows(requestingAccount, relevantAccounts.ReplyToAccount)
			if err != nil {
				return false, err
			}
			if !followsRepliedAccount {
				l.Trace("target status is a followers-only reply to an account that is not followed by the requesting account")
				return false, nil
			}
		}
	}

	// status boosts accounts id
	if relevantAccounts.BoostedStatusAuthor != nil {
		if blocked, err := f.db.Blocked(relevantAccounts.BoostedStatusAuthor.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and boosted account")
			return false, nil
		}
	}

	// status boosts a reply to account id
	if relevantAccounts.BoostedReplyToAccount != nil {
		if blocked, err := f.db.Blocked(relevantAccounts.BoostedReplyToAccount.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and boosted reply to account")
			return false, nil
		}
	}

	// status mentions accounts
	for _, a := range relevantAccounts.MentionedAccounts {
		if blocked, err := f.db.Blocked(a.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and a mentioned account")
			return false, nil
		}
	}

	// boost mentions accounts
	for _, a := range relevantAccounts.BoostedMentionedAccounts {
		if blocked, err := f.db.Blocked(a.ID, requestingAccount.ID, true); err != nil {
			return false, err
		} else if blocked {
			l.Trace("a block exists between requesting account and a boosted mentioned account")
			return false, nil
		}
	}

	// if the requesting account is mentioned in the status it should always be visible
	for _, acct := range relevantAccounts.MentionedAccounts {
		if acct.ID == requestingAccount.ID {
			return true, nil // yep it's mentioned!
		}
	}

	// at this point we know neither account blocks the other, or another account mentioned or otherwise referred to in the status
	// that means it's now just a matter of checking the visibility settings of the status itself
	switch targetStatus.Visibility {
	case gtsmodel.VisibilityPublic, gtsmodel.VisibilityUnlocked:
		// no problem here, just return OK
		return true, nil
	case gtsmodel.VisibilityFollowersOnly:
		// check one-way follow
		follows, err := f.db.Follows(requestingAccount, targetAccount)
		if err != nil {
			return false, err
		}
		if !follows {
			l.Trace("requested status is followers only but requesting account is not a follower")
			return false, nil
		}
		return true, nil
	case gtsmodel.VisibilityMutualsOnly:
		// check mutual follow
		mutuals, err := f.db.Mutuals(requestingAccount, targetAccount)
		if err != nil {
			return false, err
		}
		if !mutuals {
			l.Trace("requested status is mutuals only but accounts aren't mufos")
			return false, nil
		}
		return true, nil
	case gtsmodel.VisibilityDirect:
		l.Trace("requesting account requests a status it's not mentioned in")
		return false, nil // it's not mentioned -_-
	}

	return false, errors.New("reached the end of StatusVisible with no result")
}
