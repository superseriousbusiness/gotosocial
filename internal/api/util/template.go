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

package util

import (
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// WebPage encapsulates variables for
// rendering an HTML template within
// a standard GtS "page" template.
type WebPage struct {
	// Name of the template for rendering
	// the page. Eg., "example.tmpl".
	Template string

	// Instance model for rendering header,
	// footer, and "about" information.
	Instance *apimodel.InstanceV1

	// OGMeta for rendering page
	// "meta:og*" tags. Can be nil.
	OGMeta *OGMeta

	// Paths to CSS files to add to
	// the page as "stylesheet" entries.
	// Can be nil.
	Stylesheets []string

	// JS files to add to the
	// page as "script" entries.
	// Can be nil.
	Javascript []JavascriptEntry

	// Extra parameters to pass to
	// the template for rendering,
	// eg., "account": *Account etc.
	// Can be nil.
	Extra map[string]any
}

type JavascriptEntry struct {
	// Insert <script> tag at the end
	// of <body> rather than in <head>.
	Bottom bool

	// Path to the js file.
	Src string

	// Use async="" attribute.
	Async bool

	// Use defer="" attribute.
	Defer bool
}

// TemplateWebPage renders the given HTML template and
// page params within the standard GtS "page" template.
//
// ogMeta, stylesheets, javascript, and any extra
// properties will be provided to the template if
// set, but can all be nil.
//
// TemplateWebPage also checks whether the requesting
// clientIP is 127.0.0.1 or within a private IP range.
// If so, it injects a suggestion into the page header
// about setting trusted-proxies correctly.
func TemplateWebPage(
	c *gin.Context,
	page WebPage,
) {
	obj := map[string]any{
		"instance":    page.Instance,
		"ogMeta":      page.OGMeta,
		"stylesheets": page.Stylesheets,
		"javascript":  page.Javascript,
	}

	// Add extras to template object.
	for k, v := range page.Extra {
		obj[k] = v
	}

	// Inject trustedProxiesRec to template
	// object (or noop if not necessary).
	injectTrustedProxiesRec(c, obj)

	templatePage(c, page.Template, http.StatusOK, obj)
}

func injectTrustedProxiesRec(
	c *gin.Context,
	obj map[string]any,
) {
	if config.GetAdvancedRateLimitRequests() <= 0 {
		// If rate limiting is disabled entirely
		// there's no point in giving a trusted
		// proxies rec, as proper clientIP is
		// basically only used for rate limiting.
		return
	}

	// clientIP = the client IP that gin
	// derives based on x-forwarded-for
	// and current trusted proxies.
	clientIP := c.ClientIP()
	if clientIP == "127.0.0.1" {
		// Suggest precise 127.0.0.1/32.
		trustedProxiesRec := clientIP + "/32"
		obj["trustedProxiesRec"] = trustedProxiesRec
		return
	}

	// True if "X-Forwarded-For"
	// or "X-Real-IP" were set.
	var hasRemoteIPHeader bool
	for _, k := range []string{
		"X-Forwarded-For",
		"X-Real-IP",
	} {
		if v := c.GetHeader(k); v != "" {
			hasRemoteIPHeader = true
			break
		}
	}

	if !hasRemoteIPHeader {
		// Upstream hasn't set a
		// remote IP header so we're
		// probably not in a reverse
		// proxy setup, bail.
		return
	}

	ip, err := netip.ParseAddr(clientIP)
	if err != nil {
		log.Warnf(
			c.Request.Context(),
			"gin returned invalid clientIP %s: %v",
			clientIP, err,
		)
	}

	if !ip.IsPrivate() {
		// Upstream set a remote IP
		// header but final clientIP
		// isn't private, so upstream
		// is probably already trusted.
		// Don't inject suggestion.
		return
	}

	except := config.GetAdvancedRateLimitExceptions()
	for _, prefix := range except {
		if prefix.Contains(ip) {
			// This ip is exempt from
			// rate limiting, so don't
			// inject the suggestion.
			return
		}
	}

	// Private IP, guess if Docker.
	if DockerSubnet.Contains(ip) {
		// Suggest a CIDR that likely
		// covers this Docker subnet,
		// eg., 172.17.0.0 -> 172.17.255.255.
		trustedProxiesRec := clientIP + "/16"
		obj["trustedProxiesRec"] = trustedProxiesRec
		return
	}

	// Private IP but we don't know
	// what it is. Suggest precise CIDR.
	trustedProxiesRec := clientIP + "/32"
	obj["trustedProxiesRec"] = trustedProxiesRec
}

// DockerSubnet is a CIDR that lets one make hazy guesses
// as to whether an address is within the ranges Docker
// uses for subnets, ie., 172.16.0.0 -> 172.31.255.255.
var DockerSubnet = netip.MustParsePrefix("172.16.0.0/12")

// templateErrorPage renders the given
// HTTP code, error, and request ID
// within the standard error template.
func templateErrorPage(
	c *gin.Context,
	instance *apimodel.InstanceV1,
	code int,
	err string,
	requestID string,
) {
	const errorTmpl = "error.tmpl"

	obj := map[string]any{
		"instance":  instance,
		"code":      code,
		"error":     err,
		"requestID": requestID,
	}

	templatePage(c, errorTmpl, code, obj)
}

// template404Page renders
// a standard 404 page.
func template404Page(
	c *gin.Context,
	instance *apimodel.InstanceV1,
	requestID string,
) {
	const notFoundTmpl = "404.tmpl"

	obj := map[string]any{
		"instance":  instance,
		"requestID": requestID,
	}

	templatePage(c, notFoundTmpl, http.StatusNotFound, obj)
}

// render the given template inside
// "page.tmpl" with the provided
// code and template object.
func templatePage(
	c *gin.Context,
	template string,
	code int,
	obj map[string]any,
) {
	const pageTmpl = "page.tmpl"

	// Render given template inside the page.
	obj["pageContent"] = template

	// Inject specific page class by trimming
	// ".tmpl" suffix. In the page template
	// (see page.tmpl) this will be appended
	// with "-page", so "index.tmpl" for example
	// ends up with class "page index-page".
	obj["pageClass"] = template[:len(template)-5]

	c.HTML(code, pageTmpl, obj)
}
