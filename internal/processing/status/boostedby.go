package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) BoostedBy(account *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusBoostedBy")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error fetching status %s: %s", targetStatusID, err))
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error fetching target account %s: %s", targetStatus.AccountID, err))
	}

	l.Trace("going to see if status is visible")
	visible, err := p.filter.StatusVisible(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error seeing if status %s is visible: %s", targetStatus.ID, err))
	}

	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("StatusBoostedBy: status is not visible"))
	}

	// get ALL accounts that faved a status -- doesn't take account of blocks and mutes and stuff
	favingAccounts, err := p.db.WhoBoostedStatus(targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error seeing who boosted status: %s", err))
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, acc := range favingAccounts {
		blocked, err := p.db.Blocked(account.ID, acc.ID)
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
