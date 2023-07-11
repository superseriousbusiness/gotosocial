package account

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// PutNote updates the requesting account's private note on the target account.
func (p *Processor) PutNote(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, comment string) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, errWithCode := p.Get(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	note := &gtsmodel.Note{
		ID:              id.NewULID(),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
		Comment:         comment,
	}
	err := p.state.DB.PutNote(ctx, note)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.RelationshipGet(ctx, requestingAccount, targetAccount.ID)
}
