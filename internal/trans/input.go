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

package trans

import (
	"context"
	"errors"
	"fmt"

	transmodel "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

func (i *importer) inputEntry(ctx context.Context, entry transmodel.TransEntry) error {
	t, ok := entry[transmodel.TypeKey].(string)
	if !ok {
		return errors.New("inputEntry: could not derive entry type: missing or malformed 'type' key in json")
	}

	switch transmodel.TransType(t) {
	case transmodel.TransAccount:
		account, err := i.accountDecode(entry)
		if err != nil {
			return fmt.Errorf("inputEntry: error decoding entry into account: %s", err)
		}
		if err := i.putInDB(ctx, account); err != nil {
			return fmt.Errorf("inputEntry: error adding account to database: %s", err)
		}
		i.log.Infof("inputEntry: added account with id %s", account.ID)
		return nil
	}

	return fmt.Errorf("inputEntry: didn't recognize transtype %s", t)
}

func (i *importer) putInDB(ctx context.Context, entry interface{}) error {
	return i.db.Put(ctx, entry)
}
