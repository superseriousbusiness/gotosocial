package status

import (
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Get(requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
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

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return mastoStatus, nil
}
