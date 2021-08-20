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

package dereferencing

import (
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (d *deref) blockedDomain(host string) (bool, error) {
	b := &gtsmodel.DomainBlock{}
	err := d.db.GetWhere([]db.Where{{Key: "domain", Value: host, CaseInsensitive: true}}, b)
	if err == nil {
		// block exists
		return true, nil
	}

	if err == db.ErrNoEntries {
		// there are no entries so there's no block
		return false, nil
	}

	// there's an actual error
	return false, err
}
