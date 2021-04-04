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

type Tag struct {
	ID                string    `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	Name              string    `pg:"unique,notnull"`
	CreatedAt         time.Time `pg:"type:timestamp,notnull,default:now()"`
	UpdatedAt         time.Time `pg:"type:timestamp,notnull,default:now()"`
	Useable           bool
	Trendable         bool
	Listable          bool
	ReviewedAt        time.Time
	RequestedReviewAt time.Time
	LastStatusAt      time.Time
	MaxScore          float32
	MaxScoreAt        time.Time
}
