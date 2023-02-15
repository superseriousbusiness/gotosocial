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

package fedi

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
)

type FediProcessor struct { //nolint:revive
	db        db.DB
	federator federation.Federator
	tc        typeutils.TypeConverter
	filter    visibility.Filter
}

// New returns a new fedi processor.
func New(db db.DB, tc typeutils.TypeConverter, federator federation.Federator) FediProcessor {
	return FediProcessor{
		db:        db,
		federator: federator,
		tc:        tc,
		filter:    visibility.NewFilter(db),
	}
}
