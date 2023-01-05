/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package status

import (
	"context"
	"errors"
	"fmt"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Context(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	visible, err := p.filter.StatusVisible(ctx, targetStatus, requestingAccount)
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

	parents, err := p.db.GetStatusParents(ctx, targetStatus, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range parents {
		if v, err := p.filter.StatusVisible(ctx, status, requestingAccount); err == nil && v {
			apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
			if err == nil {
				context.Ancestors = append(context.Ancestors, *apiStatus)
			}
		}
	}

	sort.Slice(context.Ancestors, func(i int, j int) bool {
		return context.Ancestors[i].ID < context.Ancestors[j].ID
	})

	children, err := p.db.GetStatusChildren(ctx, targetStatus, false, "")
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range children {
		if v, err := p.filter.StatusVisible(ctx, status, requestingAccount); err == nil && v {
			apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
			if err == nil {
				context.Descendants = append(context.Descendants, *apiStatus)
			}
		}
	}

	return context, nil
}
