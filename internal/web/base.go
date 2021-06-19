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

package base

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

type Module sruct {
	config *config.Config
	log    *logrus.Logger
}

func New(config *config.Config, log *logrus.Logger) {
	return &Module{
		config: config,
		log: log
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	l := m.log.WithField("func", "BaseGETHandler")

	// serve static files from /assets
	assetPath := filepath.Join(cwd, m.config.TemplateConfig.AssetBaseDir);
	s.Static("/assets", assetPath);

	// serve front-page
	s.GET("/", func(c *gin.Context) {
		l.Trace("serving index html");
		// FIXME: actual variables
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"instancename":  "GoToSocial Test Instance",
			"countUsers":    3,
			"countStatuses": 42069,
			"version":       "1.0.0",
			"adminUsername": "@admin",
		});
	})
	return nil
}