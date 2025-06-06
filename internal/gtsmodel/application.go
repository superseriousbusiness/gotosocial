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

package gtsmodel

import "strings"

// Application represents an application that
// can perform actions on behalf of a user.
//
// It is equivalent to an OAuth client.
type Application struct {
	ID              string   `bun:"type:CHAR(26),pk,nullzero,notnull,unique"` // id of this item in the database
	Name            string   `bun:",notnull"`                                 // name of the application given when it was created (eg., 'tusky')
	Website         string   `bun:",nullzero"`                                // website for the application given when it was created (eg., 'https://tusky.app')
	RedirectURIs    []string `bun:"redirect_uris,array"`                      // redirect uris requested by the application for oauth2 flow
	ClientID        string   `bun:"type:CHAR(26),nullzero,notnull"`           // id of the associated oauth client entity in the db
	ClientSecret    string   `bun:",nullzero,notnull"`                        // secret of the associated oauth client entity in the db
	Scopes          string   `bun:",notnull"`                                 // scopes requested when this app was created
	ManagedByUserID string   `bun:"type:CHAR(26),nullzero"`                   // id of the user that manages this application, if it was created through the settings panel
}

// Implements oauth2.ClientInfo.
func (a *Application) GetID() string {
	return a.ClientID
}

// Implements oauth2.ClientInfo.
func (a *Application) GetSecret() string {
	return a.ClientSecret
}

// Implements oauth2.ClientInfo.
func (a *Application) GetDomain() string {
	return strings.Join(a.RedirectURIs, "\n")
}

// Implements oauth2.ClientInfo.
func (a *Application) GetUserID() string {
	return a.ManagedByUserID
}

// Implements oauth2.IsPublic.
func (a *Application) IsPublic() bool {

	// this maintains behaviour with the
	// previous version of oauth2 lib.
	return false
}
