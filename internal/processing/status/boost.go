package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Boost(requestingAccount *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
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
		if !targetStatus.VisibilityAdvanced.Boostable {
			return nil, gtserror.NewErrorForbidden(errors.New("status is not boostable"))
		}
	}

	// it's visible! it's boostable! so let's boost the FUCK out of it
	boostWrapperStatus, err := p.tc.StatusToBoost(targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	boostWrapperStatus.CreatedWithApplicationID = application.ID
	boostWrapperStatus.BoostOfAccount = targetStatus.Account

	// put the boost in the database
	if err := p.db.PutStatus(boostWrapperStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the processor for async processing
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsAnnounce,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       boostWrapperStatus,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetStatus.Account,
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := p.tc.StatusToMasto(boostWrapperStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
