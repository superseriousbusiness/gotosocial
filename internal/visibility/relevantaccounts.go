/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// relevantAccounts denotes accounts that are replied to, boosted by, or mentioned in a status.
type relevantAccounts struct {
	// Who wrote the status
	Account *gtsmodel.Account
	// Who is the status replying to
	InReplyToAccount *gtsmodel.Account
	// Which accounts are mentioned (tagged) in the status
	MentionedAccounts []*gtsmodel.Account
	// Who authed the boosted status
	BoostedAccount *gtsmodel.Account
	// If the boosted status replies to another account, who does it reply to?
	BoostedInReplyToAccount *gtsmodel.Account
	// Who is mentioned (tagged) in the boosted status
	BoostedMentionedAccounts []*gtsmodel.Account
}

func (f *filter) relevantAccounts(ctx context.Context, status *gtsmodel.Status, getBoosted bool) (*relevantAccounts, error) {
	relAccts := &relevantAccounts{
		MentionedAccounts:        []*gtsmodel.Account{},
		BoostedMentionedAccounts: []*gtsmodel.Account{},
	}

	/*
		Here's what we need to try and extract from the status:

			// 1. Who wrote the status
		    Account *gtsmodel.Account

		    // 2. Who is the status replying to
		    InReplyToAccount *gtsmodel.Account

		    // 3. Which accounts are mentioned (tagged) in the status
		    MentionedAccounts []*gtsmodel.Account

			if getBoosted:
				// 4. Who wrote the boosted status
				BoostedAccount *gtsmodel.Account

				// 5. If the boosted status replies to another account, who does it reply to?
				BoostedInReplyToAccount *gtsmodel.Account

				// 6. Who is mentioned (tagged) in the boosted status
				BoostedMentionedAccounts []*gtsmodel.Account
	*/

	// 1. Account.
	// Account might be set on the status already
	if status.Account != nil {
		// it was set
		relAccts.Account = status.Account
	} else {
		// it wasn't set, so get it from the db
		account, err := f.db.GetAccountByID(ctx, status.AccountID)
		if err != nil {
			return nil, fmt.Errorf("relevantAccounts: error getting account with id %s: %s", status.AccountID, err)
		}
		// set it on the status in case we need it further along
		status.Account = account
		// set it on relevant accounts
		relAccts.Account = account
	}

	// 2. InReplyToAccount
	// only get this if InReplyToAccountID is set
	if status.InReplyToAccountID != "" {
		// InReplyToAccount might be set on the status already
		if status.InReplyToAccount != nil {
			// it was set
			relAccts.InReplyToAccount = status.InReplyToAccount
		} else {
			// it wasn't set, so get it from the db
			inReplyToAccount, err := f.db.GetAccountByID(ctx, status.InReplyToAccountID)
			if err != nil {
				return nil, fmt.Errorf("relevantAccounts: error getting inReplyToAccount with id %s: %s", status.InReplyToAccountID, err)
			}
			// set it on the status in case we need it further along
			status.InReplyToAccount = inReplyToAccount
			// set it on relevant accounts
			relAccts.InReplyToAccount = inReplyToAccount
		}
	}

	// 3. MentionedAccounts
	// First check if status.Mentions is populated with all mentions that correspond to status.MentionIDs
	for _, mID := range status.MentionIDs {
		if mID == "" {
			continue
		}
		if !idIn(mID, status.Mentions) {
			// mention with ID isn't in status.Mentions
			mention, err := f.db.GetMention(ctx, mID)
			if err != nil {
				return nil, fmt.Errorf("relevantAccounts: error getting mention with id %s: %s", mID, err)
			}
			if mention == nil {
				return nil, fmt.Errorf("relevantAccounts: mention with id %s was nil", mID)
			}
			status.Mentions = append(status.Mentions, mention)
		}
	}
	// now filter mentions to make sure we only have mentions with a corresponding ID
	nm := []*gtsmodel.Mention{}
	for _, m := range status.Mentions {
		if m == nil {
			continue
		}
		if mentionIn(m, status.MentionIDs) {
			nm = append(nm, m)
			relAccts.MentionedAccounts = append(relAccts.MentionedAccounts, m.TargetAccount)
		}
	}
	status.Mentions = nm

	if len(status.Mentions) != len(status.MentionIDs) {
		return nil, errors.New("relevantAccounts: mentions length did not correspond with mentionIDs length")
	}

	// if getBoosted is set, we should check the same properties on the boosted account as well
	if getBoosted {
		// 4, 5, 6. Boosted status items
		// get the boosted status if it's not set on the status already
		if status.BoostOfID != "" && status.BoostOf == nil {
			boostedStatus, err := f.db.GetStatusByID(ctx, status.BoostOfID)
			if err != nil {
				return nil, fmt.Errorf("relevantAccounts: error getting boosted status with id %s: %s", status.BoostOfID, err)
			}
			status.BoostOf = boostedStatus
		}

		if status.BoostOf != nil {
			// return relevant accounts for the boosted status
			boostedRelAccts, err := f.relevantAccounts(ctx, status.BoostOf, false) // false because we don't want to recurse
			if err != nil {
				return nil, fmt.Errorf("relevantAccounts: error getting relevant accounts of boosted status %s: %s", status.BoostOf.ID, err)
			}
			relAccts.BoostedAccount = boostedRelAccts.Account
			relAccts.BoostedInReplyToAccount = boostedRelAccts.InReplyToAccount
			relAccts.BoostedMentionedAccounts = boostedRelAccts.MentionedAccounts
		}
	}

	return relAccts, nil
}

// domainBlockedRelevant checks through all relevant accounts attached to a status
// to make sure none of them are domain blocked by this instance.
func (f *filter) domainBlockedRelevant(ctx context.Context, r *relevantAccounts) (bool, error) {
	domains := []string{}

	if r.Account != nil {
		domains = append(domains, r.Account.Domain)
	}

	if r.InReplyToAccount != nil {
		domains = append(domains, r.InReplyToAccount.Domain)
	}

	for _, a := range r.MentionedAccounts {
		if a != nil {
			domains = append(domains, a.Domain)
		}
	}

	if r.BoostedAccount != nil {
		domains = append(domains, r.BoostedAccount.Domain)
	}

	if r.BoostedInReplyToAccount != nil {
		domains = append(domains, r.BoostedInReplyToAccount.Domain)
	}

	for _, a := range r.BoostedMentionedAccounts {
		if a != nil {
			domains = append(domains, a.Domain)
		}
	}

	return f.db.AreDomainsBlocked(ctx, domains)
}

func idIn(id string, mentions []*gtsmodel.Mention) bool {
	for _, m := range mentions {
		if m == nil {
			continue
		}
		if m.ID == id {
			return true
		}
	}
	return false
}

func mentionIn(mention *gtsmodel.Mention, ids []string) bool {
	if mention == nil {
		return false
	}
	for _, i := range ids {
		if mention.ID == i {
			return true
		}
	}
	return false
}
