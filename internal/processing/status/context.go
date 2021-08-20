package status

import (
	"errors"
	"fmt"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Context(requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
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

	context := &apimodel.Context{
		Ancestors:   []apimodel.Status{},
		Descendants: []apimodel.Status{},
	}

	parents, err := p.db.GetStatusParents(targetStatus, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range parents {
		if v, err := p.filter.StatusVisible(status, requestingAccount); err == nil && v {
			mastoStatus, err := p.tc.StatusToMasto(status, requestingAccount)
			if err == nil {
				context.Ancestors = append(context.Ancestors, *mastoStatus)
			}
		}
	}

	sort.Slice(context.Ancestors, func(i int, j int) bool {
		return context.Ancestors[i].ID < context.Ancestors[j].ID
	})

	children, err := p.db.GetStatusChildren(targetStatus, false, "")
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range children {
		if v, err := p.filter.StatusVisible(status, requestingAccount); err == nil && v {
			mastoStatus, err := p.tc.StatusToMasto(status, requestingAccount)
			if err == nil {
				context.Descendants = append(context.Descendants, *mastoStatus)
			}
		}
	}

	return context, nil
}
