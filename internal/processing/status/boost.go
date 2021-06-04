package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Boost(account *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusBoost")

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

	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Boostable {
			return nil, gtserror.NewErrorForbidden(errors.New("status is not boostable"))
		}
	}

	// it's visible! it's boostable! so let's boost the FUCK out of it
	boostWrapperStatus, err := p.tc.StatusToBoost(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	boostWrapperStatus.CreatedWithApplicationID = application.ID
	boostWrapperStatus.GTSBoostedAccount = targetAccount

	// put the boost in the database
	if err := p.db.Put(boostWrapperStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it to the processor for async processing
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsAnnounce,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       boostWrapperStatus,
		OriginAccount:  account,
		TargetAccount:  targetAccount,
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := p.tc.StatusToMasto(boostWrapperStatus, account, account, targetAccount, nil, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
