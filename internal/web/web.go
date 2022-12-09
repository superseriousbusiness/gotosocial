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

	"codeberg.org/gruf/go-cache/v3"
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
	rssFeedPath      = profilePath + "/feed.rss"
	statusPath       = profilePath + "/statuses/:" + statusIDKey
	assetsPathPrefix = "/assets"
	distPathPrefix   = assetsPathPrefix + "/dist"
	userPanelPath    = "/settings/user"
	adminPanelPath   = "/settings/admin"

	tokenParam  = "token"
	usernameKey = "username"
	statusIDKey = "status"

	cacheControlHeader    = "Cache-Control"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	cacheControlNoCache   = "no-cache"          // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#response_directives
	ifModifiedSinceHeader = "If-Modified-Since" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	ifNoneMatchHeader     = "If-None-Match"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
	eTagHeader            = "ETag"              // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	lastModifiedHeader    = "Last-Modified"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
)

// Module implements the api.ClientModule interface for web pages.
type Module struct {
	processor processing.Processor
	eTagCache cache.Cache[string, eTagCacheEntry]
}

// New returns a new api.ClientModule for web pages.
func New(processor processing.Processor) api.ClientModule {
	return &Module{
		processor: processor,
		eTagCache: newETagCache(),
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	// serve static files from assets dir at /assets
	assetsGroup := s.AttachGroup(assetsPathPrefix)
	m.mountAssetsFilesystem(assetsGroup)

	s.AttachHandler(http.MethodGet, "/settings", m.SettingsPanelHandler)
	s.AttachHandler(http.MethodGet, "/settings/*panel", m.SettingsPanelHandler)

	// User panel redirects
	// used by clients
	s.AttachHandler(http.MethodGet, "/auth/edit", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, userPanelPath)
	})

	// old version of settings panel
	s.AttachHandler(http.MethodGet, "/user", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, userPanelPath)
	})

	// Admin panel redirects
	// old version of settings panel
	s.AttachHandler(http.MethodGet, "/admin", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, adminPanelPath)
	})

	// serve front-page
	s.AttachHandler(http.MethodGet, "/", m.baseHandler)

	// serve profile pages at /@username
	s.AttachHandler(http.MethodGet, profilePath, m.profileGETHandler)

	// serve custom css at /@username/custom.css
	s.AttachHandler(http.MethodGet, customCSSPath, m.customCSSGETHandler)

	s.AttachHandler(http.MethodGet, rssFeedPath, m.rssFeedGETHandler)

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
