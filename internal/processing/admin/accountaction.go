package admin

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *processor) AccountAction(ctx context.Context, account *gtsmodel.Account, form *apimodel.AdminAccountActionRequest) gtserror.WithCode {
	targetAccount, err := p.db.GetAccountByID(ctx, form.TargetAccountID)
	if err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	adminAction := &gtsmodel.AdminAccountAction{
		ID:              id.NewULID(),
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
		Text:            form.Text,
	}

	switch form.Type {
	case string(gtsmodel.AdminActionSuspend):
		adminAction.Type = gtsmodel.AdminActionSuspend
		// pass the account delete through the client api channel for processing
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			OriginAccount:  account,
			TargetAccount:  targetAccount,
		})
	default:
		return gtserror.NewErrorBadRequest(fmt.Errorf("admin action type %s is not supported for this endpoint", form.Type))
	}

	if err := p.db.Put(ctx, adminAction); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
