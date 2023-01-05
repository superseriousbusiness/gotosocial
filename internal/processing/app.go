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

package processing

import (
	"context"

	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AppCreate(ctx context.Context, authed *oauth.Auth, form *apimodel.ApplicationCreateRequest) (*apimodel.Application, gtserror.WithCode) {
	// set default 'read' for scopes if it's not set
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// generate new IDs for this application and its associated client
	clientID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	clientSecret := uuid.NewString()

	appID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// generate the application to put in the database
	app := &gtsmodel.Application{
		ID:           appID,
		Name:         form.ClientName,
		Website:      form.Website,
		RedirectURI:  form.RedirectURIs,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
	}

	// chuck it in the db
	if err := p.db.Put(ctx, app); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// now we need to model an oauth client from the application that the oauth library can use
	oc := &gtsmodel.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: form.RedirectURIs,
		// This client isn't yet associated with a specific user,  it's just an app client right now
		UserID: "",
	}

	// chuck it in the db
	if err := p.db.Put(ctx, oc); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiApp, err := p.tc.AppToAPIAppSensitive(ctx, app)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiApp, nil
}
