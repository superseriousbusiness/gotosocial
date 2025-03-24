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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// Followers obtains the Followers Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	acct, err := f.state.DB.GetAccountByURI(ctx, actorIRI.String())
	if err != nil {
		return nil, err
	}

	// Fetch followers for account from database.
	follows, err := f.state.DB.GetAccountFollowers(ctx, acct.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("Followers: db error getting followers for account id %s: %s", acct.ID, err)
	}

	// Convert the followers to a slice of account URIs.
	iris := make([]*url.URL, 0, len(follows))
	for _, follow := range follows {
		u, err := url.Parse(follow.Account.URI)
		if err != nil {
			return nil, gtserror.Newf("invalid account uri: %v", err)
		}
		iris = append(iris, u)
	}

	return f.collectIRIs(ctx, iris)
}
