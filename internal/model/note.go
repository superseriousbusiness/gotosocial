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

import "time"

type Note struct {
	ID             string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	URI            string
	URL            string
	Content        string
	CreatedAt      time.Time `pg:"type:timestamp,notnull"`
	UpdatedAt      time.Time `pg:"type:timestamp,notnull"`
	Local          bool
	AccountID      string
	InReplyToID    string
	BoostOfID      string
	ContentWarning string
	Visibility     *Visibility
}

type Visibility struct {
	Direct    bool
	Followers bool
	Local     bool
	Unlisted  bool
	Public    bool
}
