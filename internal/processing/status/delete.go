package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Delete(requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	if targetStatus.AccountID != requestingAccount.ID {
		return nil, gtserror.NewErrorForbidden(errors.New("status doesn't belong to requesting account"))
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, requestingAccount)
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
		OriginAccount:  requestingAccount,
		TargetAccount:  requestingAccount,
	}

	return mastoStatus, nil
}
