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
)

func (p *Processor) Delete(
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

	// Convert app before deletion.
	apiApp, err := p.converter.AppToAPIAppSensitive(ctx, app)
	if err != nil {
		err := gtserror.Newf("error converting app to api app: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Delete app itself.
	if err := p.state.DB.DeleteApplicationByID(ctx, appID); err != nil {
		err := gtserror.Newf("db error deleting app %s: %w", appID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Delete all tokens owned by app.
	if err := p.state.DB.DeleteTokensByClientID(ctx, app.ClientID); err != nil {
		err := gtserror.Newf("db error deleting tokens for app %s: %w", appID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiApp, nil
}
