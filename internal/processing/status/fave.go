package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) Fave(requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
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
	if targetStatus.VisibilityAdvanced != nil {
		if !targetStatus.VisibilityAdvanced.Likeable {
			return nil, gtserror.NewErrorForbidden(errors.New("status is not faveable"))
		}
	}

	// first check if the status is already faved, if so we don't need to do anything
	newFave := true
	gtsFave := &gtsmodel.StatusFave{}
	if err := p.db.GetWhere([]db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: requestingAccount.ID}}, gtsFave); err == nil {
		// we already have a fave for this status
		newFave = false
	}

	if newFave {
		thisFaveID, err := id.NewULID()
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}

		// we need to create a new fave in the database
		gtsFave := &gtsmodel.StatusFave{
			ID:              thisFaveID,
			AccountID:       requestingAccount.ID,
			Account:         requestingAccount,
			TargetAccountID: targetStatus.AccountID,
			TargetAccount:   targetStatus.Account,
			StatusID:        targetStatus.ID,
			Status:          targetStatus,
			URI:             util.GenerateURIForLike(requestingAccount.Username, p.config.Protocol, p.config.Host, thisFaveID),
		}

		if err := p.db.Put(gtsFave); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting fave in database: %s", err))
		}

		// send it back to the processor for async processing
		p.fromClientAPI <- gtsmodel.FromClientAPI{
			APObjectType:   gtsmodel.ActivityStreamsLike,
			APActivityType: gtsmodel.ActivityStreamsCreate,
			GTSModel:       gtsFave,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetStatus.Account,
		}
	}

	// return the mastodon representation of the target status
	mastoStatus, err := p.tc.StatusToMasto(targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
