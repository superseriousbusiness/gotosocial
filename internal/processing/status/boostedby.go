package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) BoostedBy(requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	visible, err := p.filter.StatusVisible(targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// get ALL accounts that faved a status -- doesn't take account of blocks and mutes and stuff
	favingAccounts, err := p.db.WhoBoostedStatus(targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error seeing who boosted status: %s", err))
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, acc := range favingAccounts {
		blocked, err := p.db.Blocked(requestingAccount.ID, acc.ID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error checking blocks: %s", err))
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// now we can return the masto representation of those accounts
	mastoAccounts := []*apimodel.Account{}
	for _, acc := range filteredAccounts {
		mastoAccount, err := p.tc.AccountToMastoPublic(acc)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusFavedBy: error converting account to api model: %s", err))
		}
		mastoAccounts = append(mastoAccounts, mastoAccount)
	}

	return mastoAccounts, nil
}
