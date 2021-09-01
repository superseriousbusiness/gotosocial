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

package gtsmodel

// Application represents an application that can perform actions on behalf of a user.
// It is used to authorize tokens etc, and is associated with an oauth client id in the database.
type Application struct {
	ID           string `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull"` // id of this application in the db
	Name         string `validate:"required" bun:",nullzero,notnull"`                      // name of the application given when it was created (eg., 'tusky')
	Website      string `validate:"omitempty,url" bun:",nullzero"`                         // website for the application given when it was created (eg., 'https://tusky.app')
	RedirectURI  string `validate:"required" bun:",nullzero,notnull"`                      // redirect uri requested by the application for oauth2 flow
	ClientID     string `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`           // id of the associated oauth client entity in the db
	ClientSecret string `validate:"required,uuid" bun:",nullzero,notnull"`                 // secret of the associated oauth client entity in the db
	Scopes       string `validate:"required" bun:",nullzero,default:'read'"`               // scopes requested when this app was created
	VapidKey     string `validate:"-" bun:",nullzero"`                                     // a vapid key generated for this app when it was created
}
