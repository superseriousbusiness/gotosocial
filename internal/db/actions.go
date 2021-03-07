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

package db

import (
	"context"

	"github.com/gotosocial/gotosocial/internal/action"
	"github.com/gotosocial/gotosocial/internal/config"
	"github.com/sirupsen/logrus"
)

// Initialize will initialize the database given in the config for use with GoToSocial
var Initialize action.GTSAction = func(ctx context.Context, c *config.Config, log *logrus.Logger) error {
	db, err := New(ctx, c, log)
	if err != nil {
		return err
	}
	return db.CreateSchema(ctx)
}
