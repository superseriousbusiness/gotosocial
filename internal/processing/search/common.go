package search

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// packageAccounts is a util function that just
// converts the given accounts into an apimodel
// account slice, or errors appropriately.
func (p *Processor) packageAccounts(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	accounts []*gtsmodel.Account,
) ([]*apimodel.Account, gtserror.WithCode) {
	apiAccounts := make([]*apimodel.Account, 0, len(accounts))

	for _, account := range accounts {
		// Ensure requester can see result account.
		visible, err := p.filter.AccountVisible(ctx, requestingAccount, account)
		if err != nil {
			err = gtserror.Newf("error checking visibility of account %s for account %s: %w", account.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "account %s is not visible to account %s, skipping this result", account.ID, requestingAccount.ID)
			continue
		}

		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			log.Debugf(ctx, "skipping account %s because it couldn't be converted to its api representation: %s", account.ID, err)
			continue
		}

		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}

// packageStatuses is a util function that just
// converts the given statuses into an apimodel
// status slice, or errors appropriately.
func (p *Processor) packageStatuses(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	statuses []*gtsmodel.Status,
) ([]*apimodel.Status, gtserror.WithCode) {
	apiStatuses := make([]*apimodel.Status, 0, len(statuses))

	for _, status := range statuses {
		// Ensure requester can see result status.
		visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
		if err != nil {
			err = gtserror.Newf("error checking visibility of status %s for account %s: %w", status.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "status %s is not visible to account %s, skipping this result", status.ID, requestingAccount.ID)
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", status.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}
