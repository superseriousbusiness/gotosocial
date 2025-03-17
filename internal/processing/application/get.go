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

package application

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

func (p *Processor) Get(
	ctx context.Context,
	userID string,
	appID string,
) (*apimodel.Application, gtserror.WithCode) {
	app, err := p.state.DB.GetApplicationByID(ctx, appID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting app %s: %w", appID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if app == nil {
		err := gtserror.Newf("app %s not found in the db", appID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if app.ManagedByUserID != userID {
		err := gtserror.Newf("app %s not managed by user %s", appID, userID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	apiApp, err := p.converter.AppToAPIAppSensitive(ctx, app)
	if err != nil {
		err := gtserror.Newf("error converting app to api app: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiApp, nil
}

func (p *Processor) GetPage(
	ctx context.Context,
	userID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	apps, err := p.state.DB.GetApplicationsManagedByUserID(ctx, userID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting apps: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(apps)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Get the lowest and highest
		// ID values, used for paging.
		lo = apps[count-1].ID
		hi = apps[0].ID

		// Best-guess items length.
		items = make([]interface{}, 0, count)
	)

	for _, app := range apps {
		apiApp, err := p.converter.AppToAPIAppSensitive(ctx, app)
		if err != nil {
			log.Errorf(ctx, "error converting app to api app: %v", err)
			continue
		}

		// Append req to return items.
		items = append(items, apiApp)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/apps",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}
