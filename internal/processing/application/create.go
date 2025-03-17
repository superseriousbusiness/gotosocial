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
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *Processor) Create(
	ctx context.Context,
	managedByUserID string,
	form *apimodel.ApplicationCreateRequest,
) (*apimodel.Application, gtserror.WithCode) {
	// Set default 'read' for
	// scopes if it's not set.
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// Normalize + parse requested redirect URIs.
	form.RedirectURIs = strings.TrimSpace(form.RedirectURIs)
	var redirectURIs []string
	if form.RedirectURIs != "" {
		// Redirect URIs can be just one value, or can be passed
		// as a newline-separated list of strings. Ensure each URI
		// is parseable + normalize it by reconstructing from *url.URL.
		// Also ensure we don't add multiple copies of the same URI.
		redirectStrs := strings.Split(form.RedirectURIs, "\n")
		added := make(map[string]struct{}, len(redirectStrs))

		for _, redirectStr := range redirectStrs {
			redirectStr = strings.TrimSpace(redirectStr)
			if redirectStr == "" {
				continue
			}

			redirectURI, err := url.Parse(redirectStr)
			if err != nil {
				errText := fmt.Sprintf("error parsing redirect URI: %v", err)
				return nil, gtserror.NewErrorBadRequest(err, errText)
			}

			redirectURIStr := redirectURI.String()
			if _, alreadyAdded := added[redirectURIStr]; !alreadyAdded {
				redirectURIs = append(redirectURIs, redirectURIStr)
				added[redirectURIStr] = struct{}{}
			}
		}

		if len(redirectURIs) == 0 {
			errText := "no redirect URIs left after trimming space"
			return nil, gtserror.NewErrorBadRequest(errors.New(errText), errText)
		}
	} else {
		// No redirect URI(s) provided, just set default oob.
		redirectURIs = append(redirectURIs, oauth.OOBURI)
	}

	// Generate random client ID.
	clientID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Generate + store app
	// to put in the database.
	app := &gtsmodel.Application{
		ID:              id.NewULID(),
		Name:            form.ClientName,
		Website:         form.Website,
		RedirectURIs:    redirectURIs,
		ClientID:        clientID,
		ClientSecret:    uuid.NewString(),
		Scopes:          scopes,
		ManagedByUserID: managedByUserID,
	}
	if err := p.state.DB.PutApplication(ctx, app); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiApp, err := p.converter.AppToAPIAppSensitive(ctx, app)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiApp, nil
}
