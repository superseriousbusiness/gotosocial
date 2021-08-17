package status

import (
	"fmt"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Context(account *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {

	context := &apimodel.Context{
		Ancestors:   []apimodel.Status{},
		Descendants: []apimodel.Status{},
	}

	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	visible, err := p.filter.StatusVisible(targetStatus, account)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(err)
	}
	if !visible {
		return nil, gtserror.NewErrorForbidden(fmt.Errorf("account with id %s does not have permission to view status %s", account.ID, targetStatusID))
	}

	parents, err := p.db.StatusParents(targetStatus, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range parents {
		if v, err := p.filter.StatusVisible(status, account); err == nil && v {
			mastoStatus, err := p.tc.StatusToMasto(status, account)
			if err == nil {
				context.Ancestors = append(context.Ancestors, *mastoStatus)
			}
		}
	}

	sort.Slice(context.Ancestors, func(i int, j int) bool {
		return context.Ancestors[i].ID < context.Ancestors[j].ID
	})

	children, err := p.db.StatusChildren(targetStatus, false, "")
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range children {
		if v, err := p.filter.StatusVisible(status, account); err == nil && v {
			mastoStatus, err := p.tc.StatusToMasto(status, account)
			if err == nil {
				context.Descendants = append(context.Descendants, *mastoStatus)
			}
		}
	}

	return context, nil
}
