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
	"errors"
	"net/http"
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	confirmEmailPath = "/" + uris.ConfirmEmailPath
	profilePath      = "/@:" + usernameKey
	customCSSPath    = profilePath + "/custom.css"
	statusPath       = profilePath + "/statuses/:" + statusIDKey
	adminPanelPath   = "/admin"
	userPanelpath    = "/user"
	assetsPathPrefix = "/assets"

	tokenParam  = "token"
	usernameKey = "username"
	statusIDKey = "status"
)

// Module implements the api.ClientModule interface for web pages.
type Module struct {
	processor       processing.Processor
	assetsETagCache cache.Cache[string, eTagCacheEntry]
}

// New returns a new api.ClientModule for web pages.
func New(processor processing.Processor) api.ClientModule {
	assetsETagCache := cache.New[string, eTagCacheEntry]()
	assetsETagCache.SetTTL(time.Hour, false)
	assetsETagCache.Start(time.Minute)

	return &Module{
		processor:       processor,
		assetsETagCache: assetsETagCache,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	// serve static files from assets dir at /assets
	assetsGroup := s.AttachGroup(assetsPathPrefix)
	m.mountAssetsFilesystem(assetsGroup)

	s.AttachHandler(http.MethodGet, "/settings", m.SettingsPanelHandler)
	s.AttachHandler(http.MethodGet, "/settings/*panel", m.SettingsPanelHandler)

	// redirect /auth/edit to /settings/user
	s.AttachHandler(http.MethodGet, "/auth/edit", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/settings/user")
	})

	// serve front-page
	s.AttachHandler(http.MethodGet, "/", m.baseHandler)

	// serve profile pages at /@username
	s.AttachHandler(http.MethodGet, profilePath, m.profileGETHandler)

	// serve custom css at /@username/custom.css
	s.AttachHandler(http.MethodGet, customCSSPath, m.customCSSGETHandler)

	// serve statuses
	s.AttachHandler(http.MethodGet, statusPath, m.threadGETHandler)

	// serve email confirmation page at /confirm_email?token=whatever
	s.AttachHandler(http.MethodGet, confirmEmailPath, m.confirmEmailGETHandler)

	// 404 handler
	s.AttachNoRouteHandler(func(c *gin.Context) {
		api.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), m.processor.InstanceGet)
	})

	return nil
}
