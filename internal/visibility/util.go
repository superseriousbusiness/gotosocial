package visibility

import (
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) pullRelevantAccountsFromStatus(targetStatus *gtsmodel.Status) (*relevantAccounts, error) {
	accounts := &relevantAccounts{
		MentionedAccounts: []*gtsmodel.Account{},
	}

	// get the author account
	if targetStatus.GTSAuthorAccount == nil {
		statusAuthor := &gtsmodel.Account{}
		if err := f.db.GetByID(targetStatus.AccountID, statusAuthor); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting statusAuthor with id %s: %s", targetStatus.AccountID, err)
		}
		targetStatus.GTSAuthorAccount = statusAuthor
	}
	accounts.StatusAuthor = targetStatus.GTSAuthorAccount

	// get the replied to account from the status and add it to the pile
	if targetStatus.InReplyToAccountID != "" {
		repliedToAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(targetStatus.InReplyToAccountID, repliedToAccount); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting repliedToAcount with id %s: %s", targetStatus.InReplyToAccountID, err)
		}
		accounts.ReplyToAccount = repliedToAccount
	}

	// get the boosted account from the status and add it to the pile
	if targetStatus.BoostOfID != "" {
		// retrieve the boosted status first
		boostedStatus := &gtsmodel.Status{}
		if err := f.db.GetByID(targetStatus.BoostOfID, boostedStatus); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting boostedStatus with id %s: %s", targetStatus.BoostOfID, err)
		}
		boostedAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(boostedStatus.AccountID, boostedAccount); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting boostedAccount with id %s: %s", boostedStatus.AccountID, err)
		}
		accounts.BoostedAccount = boostedAccount

		// the boosted status might be a reply to another account so we should get that too
		if boostedStatus.InReplyToAccountID != "" {
			boostedStatusRepliedToAccount := &gtsmodel.Account{}
			if err := f.db.GetByID(boostedStatus.InReplyToAccountID, boostedStatusRepliedToAccount); err != nil {
				return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting boostedStatusRepliedToAccount with id %s: %s", boostedStatus.InReplyToAccountID, err)
			}
			accounts.BoostedReplyToAccount = boostedStatusRepliedToAccount
		}
	}

	// now get all accounts with IDs that are mentioned in the status
	for _, mentionID := range targetStatus.Mentions {

		mention := &gtsmodel.Mention{}
		if err := f.db.GetByID(mentionID, mention); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting mention with id %s: %s", mentionID, err)
		}

		mentionedAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(mention.TargetAccountID, mentionedAccount); err != nil {
			return accounts, fmt.Errorf("PullRelevantAccountsFromStatus: error getting mentioned account: %s", err)
		}
		accounts.MentionedAccounts = append(accounts.MentionedAccounts, mentionedAccount)
	}

	return accounts, nil
}

// relevantAccounts denotes accounts that are replied to, boosted by, or mentioned in a status.
type relevantAccounts struct {
	StatusAuthor          *gtsmodel.Account
	ReplyToAccount        *gtsmodel.Account
	BoostedAccount        *gtsmodel.Account
	BoostedReplyToAccount *gtsmodel.Account
	MentionedAccounts     []*gtsmodel.Account
}
