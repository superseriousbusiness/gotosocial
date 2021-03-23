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

package model

import "github.com/gotosocial/gotosocial/pkg/mastotypes"

// Application represents an application that can perform actions on behalf of a user.
// It is used to authorize tokens etc, and is associated with an oauth client id in the database.
type Application struct {
	// id of this application in the db
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	// name of the application given when it was created (eg., 'tusky')
	Name string
	// website for the application given when it was created (eg., 'https://tusky.app')
	Website string
	// redirect uri requested by the application for oauth2 flow
	RedirectURI string
	// id of the associated oauth client entity in the db
	ClientID string
	// secret of the associated oauth client entity in the db
	ClientSecret string
	// scopes requested when this app was created
	Scopes string
	// a vapid key generated for this app when it was created
	VapidKey string
}

// ToMasto returns this application as a mastodon api type, ready for serialization
func (a *Application) ToMasto() *mastotypes.Application {
	return &mastotypes.Application{
		ID:           a.ID,
		Name:         a.Name,
		Website:      a.Website,
		RedirectURI:  a.RedirectURI,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		VapidKey:     a.VapidKey,
	}
}
