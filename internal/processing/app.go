/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) AppCreate(authed *oauth.Auth, form *apimodel.ApplicationCreateRequest) (*apimodel.Application, error) {
	// set default 'read' for scopes if it's not set, this follows the default of the mastodon api https://docs.joinmastodon.org/methods/apps/
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// generate new IDs for this application and its associated client
	clientID := uuid.NewString()
	clientSecret := uuid.NewString()
	vapidKey := uuid.NewString()

	// generate the application to put in the database
	app := &gtsmodel.Application{
		Name:         form.ClientName,
		Website:      form.Website,
		RedirectURI:  form.RedirectURIs,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		VapidKey:     vapidKey,
	}

	// chuck it in the db
	if err := p.db.Put(app); err != nil {
		return nil, err
	}

	// now we need to model an oauth client from the application that the oauth library can use
	oc := &oauth.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: form.RedirectURIs,
		UserID: "", // This client isn't yet associated with a specific user,  it's just an app client right now
	}

	// chuck it in the db
	if err := p.db.Put(oc); err != nil {
		return nil, err
	}

	mastoApp, err := p.tc.AppToMastoSensitive(app)
	if err != nil {
		return nil, err
	}

	return mastoApp, nil
}
