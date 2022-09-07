/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (m *Module) SettingsPanelHandler(c *gin.Context) {
	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"instance": instance,
		"stylesheets": []string{
			assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
			assetsPathPrefix + "/dist/_colors.css",
			assetsPathPrefix + "/dist/base.css",
			assetsPathPrefix + "/dist/settings-panel-style.css",
		},
		"javascript": []string{
			assetsPathPrefix + "/dist/bundle.js",
			assetsPathPrefix + "/dist/settings.js",
		},
	})
}

func (m *Module) UserPanelHandler(c *gin.Context) {
	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"instance": instance,
		"stylesheets": []string{
			assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
			assetsPathPrefix + "/dist/_colors.css",
			assetsPathPrefix + "/dist/base.css",
			assetsPathPrefix + "/dist/panels-base.css",
			assetsPathPrefix + "/dist/panels-user-style.css",
		},
		"javascript": []string{
			assetsPathPrefix + "/dist/bundle.js",
			assetsPathPrefix + "/dist/user-panel.js",
		},
	})
}

// TODO: abstract the {admin, user}panel handlers in some way
func (m *Module) AdminPanelHandler(c *gin.Context) {
	host := config.GetHost()
	instance, err := m.processor.InstanceGet(c.Request.Context(), host)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.HTML(http.StatusOK, "frontend.tmpl", gin.H{
		"instance": instance,
		"stylesheets": []string{
			assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css",
			assetsPathPrefix + "/dist/_colors.css",
			assetsPathPrefix + "/dist/base.css",
			assetsPathPrefix + "/dist/panels-base.css",
			assetsPathPrefix + "/dist/panels-admin-style.css",
		},
		"javascript": []string{
			assetsPathPrefix + "/dist/bundle.js",
			assetsPathPrefix + "/dist/admin-panel.js",
		},
	})
}
