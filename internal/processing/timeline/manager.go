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

package timeline

import (
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type Manager interface {
	Ingest(status *gtsmodel.Status) error
	HomeTimelineGet(account *gtsmodel.Account, maxID string, sinceID string, minID string, limit int, local bool) ([]apimodel.Status, error)
}

func NewManager(db db.DB, config *config.Config) Manager {
	return &manager{
		accountTimelines: make(map[string]*timeline),
		db:               db,
		config:           config,
	}
}

type manager struct {
	accountTimelines map[string]*timeline
	db               db.DB
	config           *config.Config
}

func (m *manager) Ingest(status *gtsmodel.Status) error {
	return nil
}

func (m *manager) HomeTimelineGet(account *gtsmodel.Account, maxID string, sinceID string, minID string, limit int, local bool) ([]apimodel.Status, error) {
	return nil, nil
}
