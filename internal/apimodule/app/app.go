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

package app

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const appsPath = "/api/v1/apps"

type appModule struct {
	server         oauth.Server
	db             db.DB
	mastoConverter mastotypes.Converter
	log            *logrus.Logger
}

// New returns a new auth module
func New(srv oauth.Server, db db.DB, mastoConverter mastotypes.Converter, log *logrus.Logger) apimodule.ClientAPIModule {
	return &appModule{
		server:         srv,
		db:             db,
		mastoConverter: mastoConverter,
		log:            log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *appModule) Route(s router.Router) error {
	s.AttachHandler(http.MethodPost, appsPath, m.appsPOSTHandler)
	return nil
}

func (m *appModule) CreateTables(db db.DB) error {
	models := []interface{}{
		&oauth.Client{},
		&oauth.Token{},
		&gtsmodel.User{},
		&gtsmodel.Account{},
		&gtsmodel.Application{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
