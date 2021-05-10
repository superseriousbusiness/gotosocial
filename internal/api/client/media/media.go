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

package media

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/message"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

// BasePath is the base API path for making media requests
const BasePath = "/api/v1/media"
// IDKey is the key for media attachment IDs
const IDKey = "id"
// BasePathWithID corresponds to a media attachment with the given ID
const BasePathWithID = BasePath + "/:" + IDKey

// Module implements the ClientAPIModule interface for media
type Module struct {
	config    *config.Config
	processor message.Processor
	log       *logrus.Logger
}

// New returns a new auth module
func New(config *config.Config, processor message.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
		log:       log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodPost, BasePath, m.MediaCreatePOSTHandler)
	s.AttachHandler(http.MethodGet, BasePathWithID, m.MediaGETHandler)
	s.AttachHandler(http.MethodPut, BasePathWithID, m.MediaPUTHandler)
	return nil
}

// CreateTables populates necessary tables in the given DB
func (m *Module) CreateTables(db db.DB) error {
	models := []interface{}{
		&gtsmodel.MediaAttachment{},
	}

	for _, m := range models {
		if err := db.CreateTable(m); err != nil {
			return fmt.Errorf("error creating table: %s", err)
		}
	}
	return nil
}
