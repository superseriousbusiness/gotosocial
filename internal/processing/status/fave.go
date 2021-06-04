package status

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) Fave(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusFave")
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

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err))
		}
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}

	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// is the status faveable?
	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Likeable {
			return nil, gtserror.NewErrorForbidden(errors.New("status is not faveable"))
		}
	}

	// first check if the status is already faved, if so we don't need to do anything
	newFave := true
	gtsFave := &gtsmodel.StatusFave{}
	if err := p.db.GetWhere([]db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: account.ID}}, gtsFave); err == nil {
		// we already have a fave for this status
		newFave = false
	}

	if newFave {
		thisFaveID := uuid.NewString()

		// we need to create a new fave in the database
		gtsFave := &gtsmodel.StatusFave{
			ID:               thisFaveID,
			AccountID:        account.ID,
			TargetAccountID:  targetAccount.ID,
			StatusID:         targetStatus.ID,
			URI:              util.GenerateURIForLike(account.Username, p.config.Protocol, p.config.Host, thisFaveID),
			GTSStatus:        targetStatus,
			GTSTargetAccount: targetAccount,
			GTSFavingAccount: account,
		}

		if err := p.db.Put(gtsFave); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting fave in database: %s", err))
		}

		// send the new fave through the processor channel for federation etc
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsLike,
			APActivityType: gtsmodel.ActivityStreamsCreate,
			GTSModel:       gtsFave,
			OriginAccount:  account,
			TargetAccount:  targetAccount,
		}
	}

	// return the mastodon representation of the target status
	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
