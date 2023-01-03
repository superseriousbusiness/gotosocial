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
	"path/filepath"

	"codeberg.org/gruf/go-cache/v3"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	confirmEmailPath   = "/" + uris.ConfirmEmailPath
	profilePath        = "/@:" + usernameKey
	customCSSPath      = profilePath + "/custom.css"
	rssFeedPath        = profilePath + "/feed.rss"
	statusPath         = profilePath + "/statuses/:" + statusIDKey
	assetsPathPrefix   = "/assets"
	distPathPrefix     = assetsPathPrefix + "/dist"
	settingsPathPrefix = "/settings"
	settingsPanelGlob  = settingsPathPrefix + "/*panel"
	userPanelPath      = settingsPathPrefix + "/user"
	adminPanelPath     = settingsPathPrefix + "/admin"

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

type Module struct {
	processor processing.Processor
	eTagCache cache.Cache[string, eTagCacheEntry]
}

func New(processor processing.Processor) *Module {
	return &Module{
		processor: processor,
		eTagCache: newETagCache(),
	}
}

func (m *Module) Route(r router.Router, mi ...gin.HandlerFunc) {
	// serve static files from assets dir at /assets
	assetsGroup := r.AttachGroup(assetsPathPrefix)
	webAssetsAbsFilePath, err := filepath.Abs(config.GetWebAssetBaseDir())
	if err != nil {
		log.Panicf("error getting absolute path of assets dir: %s", err)
	}

	fs := fileSystem{http.Dir(webAssetsAbsFilePath)}

	// use the cache middleware on all handlers in this group
	assetsGroup.Use(m.assetsCacheControlMiddleware(fs))
	assetsGroup.Use(mi...)

	// serve static file system in the root of this group,
	// will end up being something like "/assets/"
	assetsGroup.StaticFS("/", fs)

	/*
		Attach individual web handlers which require no specific middlewares
	*/

	r.AttachHandler(http.MethodGet, "/", m.baseHandler) // front-page
	r.AttachHandler(http.MethodGet, settingsPathPrefix, m.SettingsPanelHandler)
	r.AttachHandler(http.MethodGet, settingsPanelGlob, m.SettingsPanelHandler)
	r.AttachHandler(http.MethodGet, profilePath, m.profileGETHandler)
	r.AttachHandler(http.MethodGet, customCSSPath, m.customCSSGETHandler)
	r.AttachHandler(http.MethodGet, rssFeedPath, m.rssFeedGETHandler)
	r.AttachHandler(http.MethodGet, statusPath, m.threadGETHandler)
	r.AttachHandler(http.MethodGet, confirmEmailPath, m.confirmEmailGETHandler)
	r.AttachHandler(http.MethodGet, robotsPath, m.robotsGETHandler)

	/*
		Attach redirects from old endpoints to current ones for backwards compatibility
	*/

	r.AttachHandler(http.MethodGet, "/auth/edit", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/user", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/admin", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, adminPanelPath) })
}
