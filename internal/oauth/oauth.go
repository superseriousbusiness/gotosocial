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

package oauth

import (
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/errors"
	"github.com/gotosocial/oauth2/v4/manage"
	"github.com/gotosocial/oauth2/v4/server"
	"github.com/sirupsen/logrus"
)

type API struct {
	manager *manage.Manager
	server  *server.Server
}

func New(ts oauth2.TokenStore, cs oauth2.ClientStore, log *logrus.Logger) *API {
	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)

	srv := server.NewDefaultServer(manager)
	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		log.Errorf("internal oauth error: %s", err)
		return nil
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Errorf("internal response error: %s", re.Error)
	})

	return &API{
		manager: manager,
		server:  srv,
	}
}
