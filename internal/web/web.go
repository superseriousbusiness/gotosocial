// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package web

import (
	"context"
	"net/http"
	"net/url"
	"path/filepath"

	"codeberg.org/gruf/go-cache/v3"
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

const (
	confirmEmailPath      = "/" + uris.ConfirmEmailPath
	profileGroupPath      = "/@:username"
	statusPath            = "/statuses/:" + apiutil.WebStatusIDKey // leave out the '/@:username' prefix as this will be served within the profile group
	tagsPath              = "/tags/:" + apiutil.TagNameKey
	customCSSPath         = profileGroupPath + "/custom.css"
	instanceCustomCSSPath = "/custom.css"
	rssFeedPath           = profileGroupPath + "/feed.rss"
	assetsPathPrefix      = "/assets"
	distPathPrefix        = assetsPathPrefix + "/dist"
	themesPathPrefix      = assetsPathPrefix + "/themes"
	settingsPathPrefix    = "/settings"
	settingsPanelGlob     = settingsPathPrefix + "/*panel"
	userPanelPath         = settingsPathPrefix + "/user"
	adminPanelPath        = settingsPathPrefix + "/admin"
	signupPath            = "/signup"

	cacheControlHeader    = "Cache-Control"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	cacheControlNoCache   = "no-cache"          // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#response_directives
	ifModifiedSinceHeader = "If-Modified-Since" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	ifNoneMatchHeader     = "If-None-Match"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-None-Match
	eTagHeader            = "ETag"              // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
	lastModifiedHeader    = "Last-Modified"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified

	cssFA        = assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css"
	cssAbout     = distPathPrefix + "/about.css"
	cssIndex     = distPathPrefix + "/index.css"
	cssLoginInfo = distPathPrefix + "/login-info.css"
	cssStatus    = distPathPrefix + "/status.css"
	cssThread    = distPathPrefix + "/thread.css"
	cssProfile   = distPathPrefix + "/profile.css"
	cssSettings  = distPathPrefix + "/settings-style.css"
	cssTag       = distPathPrefix + "/tag.css"

	jsFrontend = distPathPrefix + "/frontend.js" // Progressive enhancement frontend JS.
	jsSettings = distPathPrefix + "/settings.js" // Settings panel React application.
)

type Module struct {
	processor    *processing.Processor
	eTagCache    cache.Cache[string, eTagCacheEntry]
	isURIBlocked func(context.Context, *url.URL) (bool, error)
}

func New(db db.DB, processor *processing.Processor) *Module {
	return &Module{
		processor:    processor,
		eTagCache:    newETagCache(),
		isURIBlocked: db.IsURIBlocked,
	}
}

func (m *Module) Route(r *router.Router, mi ...gin.HandlerFunc) {
	// Group all static files from assets dir at /assets,
	// so that they can use the same cache control middleware.
	webAssetsAbsFilePath, err := filepath.Abs(config.GetWebAssetBaseDir())
	if err != nil {
		log.Panicf(nil, "error getting absolute path of assets dir: %s", err)
	}
	fs := fileSystem{http.Dir(webAssetsAbsFilePath)}
	assetsGroup := r.AttachGroup(assetsPathPrefix)
	assetsGroup.Use(m.assetsCacheControlMiddleware(fs))
	assetsGroup.Use(mi...)
	assetsGroup.StaticFS("/", fs)

	// handlers that serve profiles and statuses should use the SignatureCheck
	// middleware, so that requests with content-type application/activity+json
	// can still be served
	profileGroup := r.AttachGroup(profileGroupPath)
	profileGroup.Use(mi...)
	profileGroup.Use(middleware.SignatureCheck(m.isURIBlocked), middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"no-store"},
	}))
	profileGroup.Handle(http.MethodGet, "", m.profileGETHandler) // use empty path here since it's the base of the group
	profileGroup.Handle(http.MethodGet, statusPath, m.threadGETHandler)

	// Attach individual web handlers which require no specific middlewares
	r.AttachHandler(http.MethodGet, "/", m.indexHandler) // front-page
	r.AttachHandler(http.MethodGet, settingsPathPrefix, m.SettingsPanelHandler)
	r.AttachHandler(http.MethodGet, settingsPanelGlob, m.SettingsPanelHandler)
	r.AttachHandler(http.MethodGet, customCSSPath, m.customCSSGETHandler)
	r.AttachHandler(http.MethodGet, instanceCustomCSSPath, m.instanceCustomCSSGETHandler)
	r.AttachHandler(http.MethodGet, rssFeedPath, m.rssFeedGETHandler)
	r.AttachHandler(http.MethodGet, confirmEmailPath, m.confirmEmailGETHandler)
	r.AttachHandler(http.MethodPost, confirmEmailPath, m.confirmEmailPOSTHandler)
	r.AttachHandler(http.MethodGet, robotsPath, m.robotsGETHandler)
	r.AttachHandler(http.MethodGet, aboutPath, m.aboutGETHandler)
	r.AttachHandler(http.MethodGet, loginPath, m.loginGETHandler)
	r.AttachHandler(http.MethodGet, domainBlockListPath, m.domainBlockListGETHandler)
	r.AttachHandler(http.MethodGet, tagsPath, m.tagGETHandler)
	r.AttachHandler(http.MethodGet, signupPath, m.signupGETHandler)
	r.AttachHandler(http.MethodPost, signupPath, m.signupPOSTHandler)

	// Attach redirects from old endpoints to current ones for backwards compatibility
	r.AttachHandler(http.MethodGet, "/auth/edit", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/user", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/admin", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, adminPanelPath) })
}
