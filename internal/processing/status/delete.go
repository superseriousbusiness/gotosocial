package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Delete(account *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	l := p.log.WithField("func", "StatusDelete")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}

	if targetStatus.AccountID != account.ID {
		return nil, gtserror.NewErrorForbidden(errors.New("status doesn't belong to requesting account"))
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

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, account, account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	if err := p.db.DeleteByID(targetStatus.ID, targetStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error deleting status from the database: %s", err))
	}

	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsNote,
		APActivityType: gtsmodel.ActivityStreamsDelete,
		GTSModel:       targetStatus,
		OriginAccount:  account,
	}

	return mastoStatus, nil
}
