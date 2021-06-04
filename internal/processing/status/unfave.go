package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Unfave(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusUnfave")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err))
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err))
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}

	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// check if we actually have a fave for this status
	var toUnfave bool

	gtsFave := &gtsmodel.StatusFave{}
	err = p.db.GetWhere([]db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: account.ID}}, gtsFave)
	if err == nil {
		// we have a fave
		toUnfave = true
	}
	if err != nil {
		// something went wrong in the db finding the fave
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching existing fave from database: %s", err))
		}
		// we just don't have a fave
		toUnfave = false
	}

	if toUnfave {
		// we had a fave, so take some action to get rid of it
		_, err = p.db.UnfaveStatus(targetStatus, account.ID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error unfaveing status: %s", err))
		}

		// send the unfave through the processor channel for federation etc
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsLike,
			APActivityType: gtsmodel.ActivityStreamsUndo,
			GTSModel:       gtsFave,
			OriginAccount:  account,
			TargetAccount:  targetAccount,
		}
	}

	// return the status (whatever its state) back to the caller
	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err))
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
