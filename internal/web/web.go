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

	"codeberg.org/gruf/go-cache/v3"
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
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

	cssFA             = assetsPathPrefix + "/Fork-Awesome/css/fork-awesome.min.css"
	cssAbout          = distPathPrefix + "/about.css"
	cssIndex          = distPathPrefix + "/index.css"
	cssLoginInfo      = distPathPrefix + "/login-info.css"
	cssStatus         = distPathPrefix + "/status.css"
	cssThread         = distPathPrefix + "/thread.css"
	cssProfile        = distPathPrefix + "/profile.css"
	cssProfileGallery = distPathPrefix + "/profile-gallery.css"
	cssSettings       = distPathPrefix + "/settings-style.css"
	cssTag            = distPathPrefix + "/tag.css"

	jsFrontend          = distPathPrefix + "/frontend.js"           // Progressive enhancement frontend JS.
	jsFrontendPrerender = distPathPrefix + "/frontend_prerender.js" // Frontend JS that should run before page renders.
	jsSettings          = distPathPrefix + "/settings.js"           // Settings panel React application.
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

// ETagCache implements withETagCache.
func (m *Module) ETagCache() cache.Cache[string, eTagCacheEntry] {
	return m.eTagCache
}

// Route attaches the assets filesystem and profile,
// status, and other web handlers to the router.
func (m *Module) Route(r *router.Router, mi ...gin.HandlerFunc) {
	// Route static assets.
	routeAssets(m, r, mi...)

	// Handlers that serve profiles and statuses should use
	// the SignatureCheck middleware, so that requests with
	// content-type application/activity+json can be served
	profileGroup := r.AttachGroup(profileGroupPath)
	profileGroup.Use(mi...)
	profileGroup.Use(middleware.SignatureCheck(m.isURIBlocked), middleware.CacheControl(middleware.CacheControlConfig{
		Directives: []string{"no-store"},
	}))
	profileGroup.Handle(http.MethodGet, "", m.profileGETHandler) // use empty path here since it's the base of the group
	profileGroup.Handle(http.MethodGet, statusPath, m.threadGETHandler)

	// Group for all other web handlers.
	everythingElseGroup := r.AttachGroup("")
	everythingElseGroup.Use(mi...)
	everythingElseGroup.Handle(http.MethodGet, "/", m.indexHandler) // front-page
	everythingElseGroup.Handle(http.MethodGet, settingsPathPrefix, m.SettingsPanelHandler)
	everythingElseGroup.Handle(http.MethodGet, settingsPanelGlob, m.SettingsPanelHandler)
	everythingElseGroup.Handle(http.MethodGet, customCSSPath, m.customCSSGETHandler)
	everythingElseGroup.Handle(http.MethodGet, instanceCustomCSSPath, m.instanceCustomCSSGETHandler)
	everythingElseGroup.Handle(http.MethodGet, rssFeedPath, m.rssFeedGETHandler)
	everythingElseGroup.Handle(http.MethodGet, confirmEmailPath, m.confirmEmailGETHandler)
	everythingElseGroup.Handle(http.MethodPost, confirmEmailPath, m.confirmEmailPOSTHandler)
	everythingElseGroup.Handle(http.MethodGet, aboutPath, m.aboutGETHandler)
	everythingElseGroup.Handle(http.MethodGet, loginPath, m.loginGETHandler)
	everythingElseGroup.Handle(http.MethodGet, domainBlockListPath, m.domainBlockListGETHandler)
	everythingElseGroup.Handle(http.MethodGet, tagsPath, m.tagGETHandler)
	everythingElseGroup.Handle(http.MethodGet, signupPath, m.signupGETHandler)
	everythingElseGroup.Handle(http.MethodPost, signupPath, m.signupPOSTHandler)

	// Redirects from old endpoints for back compat.
	r.AttachHandler(http.MethodGet, "/auth/edit", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/user", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, userPanelPath) })
	r.AttachHandler(http.MethodGet, "/admin", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, adminPanelPath) })
}
