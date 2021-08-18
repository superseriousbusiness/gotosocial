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

package federatingdb

import (
	"context"
	"errors"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Get returns the database entry for the specified id.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	l := f.log.WithFields(
		logrus.Fields{
			"func": "Get",
			"id":   id.String(),
		},
	)
	l.Debug("entering GET function")

	if util.IsUserPath(id) {
		acct, err := f.db.GetAccountByURI(id.String())
		if err != nil {
			return nil, err
		}
		l.Debug("is user path! returning account")
		return f.typeConverter.AccountToAS(acct)
	}

	if util.IsFollowersPath(id) {
		acct := &gtsmodel.Account{}
		if err := f.db.GetWhere([]db.Where{{Key: "followers_uri", Value: id.String()}}, acct); err != nil {
			return nil, err
		}

		followersURI, err := url.Parse(acct.FollowersURI)
		if err != nil {
			return nil, err
		}

		return f.Followers(c, followersURI)
	}

	if util.IsFollowingPath(id) {
		acct := &gtsmodel.Account{}
		if err := f.db.GetWhere([]db.Where{{Key: "following_uri", Value: id.String()}}, acct); err != nil {
			return nil, err
		}

		followingURI, err := url.Parse(acct.FollowingURI)
		if err != nil {
			return nil, err
		}

		return f.Following(c, followingURI)
	}

	return nil, errors.New("could not get")
}
