package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Delete(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusDelete")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		if _, ok := err.(db.ErrNoEntries); !ok {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
		}
		// status is already gone
		return nil, nil
	}

	if targetStatus.AccountID != account.ID {
		return nil, gtserror.NewErrorForbidden(errors.New("status doesn't belong to requesting account"))
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err))
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	if err := p.db.DeleteByID(targetStatus.ID, &gtsmodel.Status{}); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error deleting status from the database: %s", err))
	}

	// send it back to the processor for async processing
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsNote,
		APActivityType: gtsmodel.ActivityStreamsDelete,
		GTSModel:       targetStatus,
		OriginAccount:  account,
	}

	return mastoStatus, nil
}
