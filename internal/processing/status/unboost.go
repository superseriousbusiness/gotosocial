package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Unboost(account *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "Unboost")

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

	l.Trace("going to see if status is visible")
	visible, err := p.filter.StatusVisible(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}

	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// check if we actually have a boost for this status
	var toUnboost bool

	gtsBoost := &gtsmodel.Status{}
	where := []db.Where{
		{
			Key:   "boost_of_id",
			Value: targetStatusID,
		},
		{
			Key:   "account_id",
			Value: account.ID,
		},
	}
	err = p.db.GetWhere(where, gtsBoost)
	if err == nil {
		// we have a boost
		toUnboost = true
	}

	if err != nil {
		// something went wrong in the db finding the boost
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching existing boost from database: %s", err))
		}
		// we just don't have a boost
		toUnboost = false
	}

	if toUnboost {
		// we had a boost, so take some action to get rid of it
		if err := p.db.DeleteWhere(where, &gtsmodel.Status{}); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error unboosting status: %s", err))
		}

		// pin some stuff onto the boost while we have it out of the db
		gtsBoost.BoostOf = targetStatus
		gtsBoost.BoostOf.Account = targetAccount
		gtsBoost.BoostOfAccount = targetAccount
		gtsBoost.Account = account

		// send it back to the processor for async processing
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsAnnounce,
			APActivityType: gtsmodel.ActivityStreamsUndo,
			GTSModel:       gtsBoost,
			OriginAccount:  account,
			TargetAccount:  targetAccount,
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
