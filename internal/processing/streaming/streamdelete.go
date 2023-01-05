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

package streaming

import (
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

func (p *processor) StreamDelete(statusID string) error {
	errs := []string{}

	// get all account IDs with open streams
	accountIDs := []string{}
	p.streamMap.Range(func(k interface{}, _ interface{}) bool {
		key, ok := k.(string)
		if !ok {
			panic("streamMap key was not a string (account id)")
		}

		accountIDs = append(accountIDs, key)
		return true
	})

	// stream the delete to every account
	for _, accountID := range accountIDs {
		if err := p.streamToAccount(statusID, stream.EventTypeDelete, stream.AllStatusTimelines, accountID); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("one or more errors streaming status delete: %s", strings.Join(errs, ";"))
	}

	return nil
}
