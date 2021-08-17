package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Unboost(requestingAccount *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
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
			Value: requestingAccount.ID,
		},
	}
	err = p.db.GetWhere(where, gtsBoost)
	if err == nil {
		// we have a boost
		toUnboost = true
	}

	if err != nil {
		// something went wrong in the db finding the boost
		if err != db.ErrNoEntries {
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
		gtsBoost.Account = requestingAccount

		gtsBoost.BoostOf = targetStatus
		gtsBoost.BoostOfAccount = targetStatus.Account
		gtsBoost.BoostOf.Account = targetStatus.Account

		// send it back to the processor for async processing
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsAnnounce,
			APActivityType: gtsmodel.ActivityStreamsUndo,
			GTSModel:       gtsBoost,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetStatus.Account,
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
