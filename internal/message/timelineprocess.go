package message

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) HomeTimelineGet(authed *oauth.Auth, maxID string, sinceID string, minID string, limit int, local bool) ([]apimodel.Status, ErrorWithCode) {
	statuses, err := p.db.GetHomeTimelineForAccount(authed.Account.ID, maxID, sinceID, minID, limit, local)
	if err != nil {
		return nil, NewErrorInternalError(err)
	}

	apiStatuses := []apimodel.Status{}
	for _, s := range statuses {
		targetAccount := &gtsmodel.Account{}
		if err := p.db.GetByID(s.AccountID, targetAccount); err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error getting status author: %s", err))
		}

		relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(s)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error getting relevant statuses: %s", err))
		}

		visible, err := p.db.StatusVisible(s, targetAccount, authed.Account, relevantAccounts)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error checking status visibility: %s", err))
		}
		if !visible {
			continue
		}

		var boostedStatus *gtsmodel.Status
		if s.BoostOfID != "" {
			bs := &gtsmodel.Status{}
			if err := p.db.GetByID(s.BoostOfID, bs); err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error getting boosted status: %s", err))
			}
			boostedRelevantAccounts, err := p.db.PullRelevantAccountsFromStatus(bs)
			if err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error getting relevant accounts from boosted status: %s", err))
			}

			boostedVisible, err := p.db.StatusVisible(bs, relevantAccounts.BoostedAccount, authed.Account, boostedRelevantAccounts)
			if err != nil {
				return nil, NewErrorInternalError(fmt.Errorf("error checking boosted status visibility: %s", err))
			}

			if boostedVisible {
				boostedStatus = bs
			}
		}

		apiStatus, err := p.tc.StatusToMasto(s, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostedStatus)
		if err != nil {
			return nil, NewErrorInternalError(fmt.Errorf("error converting status to masto: %s", err))
		}

		apiStatuses = append(apiStatuses, *apiStatus)
	}

	return apiStatuses, nil
}
