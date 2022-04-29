package visibility

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusBoostable(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (bool, error) {
	visible, err := f.StatusVisible(ctx, targetStatus, requestingAccount)
	if err != nil {
		return false, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}
	if !visible {
		return false, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// the original account should always be able to boost its own statuses unless they are direct messages
	if requestingAccount == targetStatus.Account && targetStatus.Visibility != gtsmodel.VisibilityDirect {
		return true, nil
	}

	return targetStatus.Boostable, nil
}
