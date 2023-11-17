// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package status

import (
	"context"
	"sort"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Get gets the given status, taking account of privacy settings and blocks etc.
func (p *Processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

// WebGet gets the given status for web use, taking account of privacy settings.
func (p *Processor) WebGet(ctx context.Context, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx, nil, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	webStatus, err := p.converter.StatusToWebStatus(ctx, targetStatus, nil)
	if err != nil {
		err = gtserror.Newf("error converting status: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return webStatus, nil
}

func (p *Processor) contextGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetStatusID string,
	convert func(context.Context, *gtsmodel.Status, *gtsmodel.Account) (*apimodel.Status, error),
) (*apimodel.Context, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	context := &apimodel.Context{
		Ancestors:   []apimodel.Status{},
		Descendants: []apimodel.Status{},
	}

	parents, err := p.state.DB.GetStatusParents(ctx, targetStatus, false)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range parents {
		if v, err := p.filter.StatusVisible(ctx, requestingAccount, status); err == nil && v {
			apiStatus, err := convert(ctx, status, requestingAccount)
			if err == nil {
				context.Ancestors = append(context.Ancestors, *apiStatus)
			}
		}
	}

	sort.Slice(context.Ancestors, func(i int, j int) bool {
		return context.Ancestors[i].ID < context.Ancestors[j].ID
	})

	children, err := p.state.DB.GetStatusChildren(ctx, targetStatus, false, "")
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, status := range children {
		if v, err := p.filter.StatusVisible(ctx, requestingAccount, status); err == nil && v {
			apiStatus, err := convert(ctx, status, requestingAccount)
			if err == nil {
				context.Descendants = append(context.Descendants, *apiStatus)
			}
		}
	}

	return context, nil
}

// ContextGet returns the context (previous and following posts) from the given status ID.
func (p *Processor) ContextGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return p.contextGet(ctx, requestingAccount, targetStatusID, p.converter.StatusToAPIStatus)
}

// WebContextGet is like ContextGet, but is explicitly
// for viewing statuses via the unauthenticated web UI.
//
// TODO: a more advanced threading model could be implemented here.
func (p *Processor) WebContextGet(ctx context.Context, targetStatusID string) (*apimodel.Context, gtserror.WithCode) {
	return p.contextGet(ctx, nil, targetStatusID, p.converter.StatusToWebStatus)
}
