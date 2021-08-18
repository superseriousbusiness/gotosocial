package visibility

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) pullRelevantAccountsFromStatus(targetStatus *gtsmodel.Status) (*relevantAccounts, error) {
	accounts := &relevantAccounts{
		MentionedAccounts:        []*gtsmodel.Account{},
		BoostedMentionedAccounts: []*gtsmodel.Account{},
	}

	// get the author account if it's not set on the status already
	if targetStatus.Account == nil {
		statusAuthor, err := f.db.GetAccountByID(targetStatus.AccountID)
		if err == nil {
			targetStatus.Account = statusAuthor
		}
	}
	accounts.StatusAuthor = targetStatus.Account

	// now get all accounts with IDs that are mentioned in the status
	if targetStatus.MentionIDs != nil && targetStatus.Mentions == nil {
		mentions, err := f.db.GetMentions(targetStatus.MentionIDs)
		if err == nil {
			targetStatus.Mentions = mentions
		}
	}

	for _, m := range targetStatus.Mentions {
		if m.TargetAccount == nil {
			t, err := f.db.GetAccountByID(m.TargetAccountID)
			if err == nil {
				m.TargetAccount = t
			}
		}
		accounts.MentionedAccounts = append(accounts.MentionedAccounts, m.TargetAccount)
	}

	// get the replied to account if it's not set on the status already
	if targetStatus.InReplyToAccountID != "" && targetStatus.InReplyToAccount == nil {
		repliedToAccount, err := f.db.GetAccountByID(targetStatus.InReplyToAccountID)
		if err == nil {
			targetStatus.InReplyToAccount = repliedToAccount
		}
	}
	accounts.ReplyToAccount = targetStatus.InReplyToAccount

	// get the boosted status if it's not set on the status already
	if targetStatus.BoostOfID != "" && targetStatus.BoostOf == nil {
		boostedStatus, err := f.db.GetStatusByID(targetStatus.BoostOfID)
		if err == nil {
			targetStatus.BoostOf = boostedStatus
		}
	}

	// get the boosted account if it's not set on the status already
	if targetStatus.BoostOfAccountID != "" && targetStatus.BoostOfAccount == nil {
		if targetStatus.BoostOf != nil && targetStatus.BoostOf.Account != nil {
			targetStatus.BoostOfAccount = targetStatus.BoostOf.Account
		} else {
			boostedAccount, err := f.db.GetAccountByID(targetStatus.BoostOfAccountID)
			if err == nil {
				targetStatus.BoostOfAccount = boostedAccount
			}
		}
	}
	accounts.BoostedStatusAuthor = targetStatus.BoostOfAccount

	if targetStatus.BoostOf != nil {
		// the boosted status might be a reply to another account so we should get that too
		if targetStatus.BoostOf.InReplyToAccountID != "" && targetStatus.BoostOf.InReplyToAccount == nil {
			boostOfInReplyToAccount, err := f.db.GetAccountByID(targetStatus.BoostOf.InReplyToAccountID)
			if err == nil {
				targetStatus.BoostOf.InReplyToAccount = boostOfInReplyToAccount
			}
		}
		accounts.BoostedReplyToAccount = targetStatus.BoostOf.InReplyToAccount

		// now get all accounts with IDs that are mentioned in the boosted status
		if targetStatus.BoostOf.MentionIDs != nil && targetStatus.BoostOf.Mentions == nil {
			mentions, err := f.db.GetMentions(targetStatus.BoostOf.MentionIDs)
			if err == nil {
				targetStatus.BoostOf.Mentions = mentions
			}
		}

		for _, m := range targetStatus.BoostOf.Mentions {
			if m.TargetAccount == nil {
				t, err := f.db.GetAccountByID(m.TargetAccountID)
				if err == nil {
					m.TargetAccount = t
				}
			}
			accounts.BoostedMentionedAccounts = append(accounts.BoostedMentionedAccounts, m.TargetAccount)
		}
	}

	return accounts, nil
}

// relevantAccounts denotes accounts that are replied to, boosted by, or mentioned in a status.
type relevantAccounts struct {
	// Who wrote the status
	StatusAuthor *gtsmodel.Account
	// Who is the status replying to
	ReplyToAccount *gtsmodel.Account
	// Which accounts are mentioned (tagged) in the status
	MentionedAccounts []*gtsmodel.Account
	// Who authed the boosted status
	BoostedStatusAuthor *gtsmodel.Account
	// If the boosted status replies to another account, who does it reply to?
	BoostedReplyToAccount *gtsmodel.Account
	// Who is mentioned (tagged) in the boosted status
	BoostedMentionedAccounts []*gtsmodel.Account
}

// blockedDomain checks whether the given domain is blocked by us or not
func (f *filter) blockedDomain(host string) (bool, error) {
	b := &gtsmodel.DomainBlock{}
	err := f.db.GetWhere([]db.Where{{Key: "domain", Value: host, CaseInsensitive: true}}, b)
	if err == nil {
		// block exists
		return true, nil
	}

	if err == db.ErrNoEntries {
		// there are no entries so there's no block
		return false, nil
	}

	// there's an actual error
	return false, err
}

// domainBlockedRelevant checks through all relevant accounts attached to a status
// to make sure none of them are domain blocked by this instance.
//
// Will return true+nil if there's a block, false+nil if there's no block, or
// an error if something goes wrong.
func (f *filter) domainBlockedRelevant(r *relevantAccounts) (bool, error) {
	if r.StatusAuthor != nil {
		b, err := f.blockedDomain(r.StatusAuthor.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	if r.ReplyToAccount != nil {
		b, err := f.blockedDomain(r.ReplyToAccount.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	for _, a := range r.MentionedAccounts {
		if a == nil {
			continue
		}
		b, err := f.blockedDomain(a.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	if r.BoostedStatusAuthor != nil {
		b, err := f.blockedDomain(r.BoostedStatusAuthor.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	if r.BoostedReplyToAccount != nil {
		b, err := f.blockedDomain(r.BoostedReplyToAccount.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	for _, a := range r.BoostedMentionedAccounts {
		if a == nil {
			continue
		}
		b, err := f.blockedDomain(a.Domain)
		if err != nil {
			return false, err
		}
		if b {
			return true, nil
		}
	}

	return false, nil
}
