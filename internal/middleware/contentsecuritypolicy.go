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

package middleware

import (
	"strings"

	"codeberg.org/gruf/go-debug"
	"github.com/gin-gonic/gin"
)

func ContentSecurityPolicy(extraURIs ...string) gin.HandlerFunc {
	csp := BuildContentSecurityPolicy(extraURIs...)

	return func(c *gin.Context) {
		// Inform the browser we only load
		// CSS/JS/media using the given policy.
		c.Header("Content-Security-Policy", csp)
	}
}

func BuildContentSecurityPolicy(extraURIs ...string) string {
	const (
		defaultSrc = "default-src"
		connectSrc = "connect-src"
		objectSrc  = "object-src"
		imgSrc     = "img-src"
		mediaSrc   = "media-src"
		frames     = "frame-ancestors"

		self = "'self'"
		none = "'none'"
		blob = "blob:"
	)

	// CSP values keyed by directive.
	values := make(map[string][]string, 5)

	/*
		default-src
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/default-src
	*/

	if !debug.DEBUG {
		// Restrictive 'self' policy
		values[defaultSrc] = []string{self}
	} else {
		// If debug is enabled, allow
		// serving things from localhost
		// as well (regardless of port).
		values[defaultSrc] = []string{
			self,
			"localhost:*",
			"ws://localhost:*",
		}
	}

	/*
		connect-src
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/connect-src
	*/

	// Restrictive default policy, but
	// include ListenBrainz API for fields.
	const listenBrains = "https://api.listenbrainz.org/1/user/"
	values[connectSrc] = append(values[defaultSrc], listenBrains) //nolint

	/*
		object-src
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/object-src
	*/

	// Disallow object-src as recommended.
	values[objectSrc] = []string{none}

	/*
		img-src
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/img-src
	*/

	// Restrictive 'self' policy,
	// include extraURIs, and 'blob:'
	// for previewing uploaded images
	// (header, avi, emojis) in settings.
	values[imgSrc] = append(
		[]string{self, blob},
		extraURIs...,
	)

	/*
		media-src
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/media-src
	*/

	// Restrictive 'self' policy,
	// include extraURIs.
	values[mediaSrc] = append(
		[]string{self},
		extraURIs...,
	)

	/*
		frame-ancestors
		https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-ancestors
	*/

	// Don't allow embedding us in an iframe
	values[frames] = []string{none}

	/*
		Assemble policy directives.
	*/

	// Iterate through an ordered slice rather than
	// iterating through the map, since we want these
	// policyDirectives in a determinate order.
	policyDirectives := make([]string, 5)
	for i, directive := range []string{
		defaultSrc,
		connectSrc,
		objectSrc,
		imgSrc,
		mediaSrc,
	} {
		// Each policy directive should look like:
		// `[directive] [value1] [value2] [etc]`

		// Get assembled values
		// for this directive.
		values := values[directive]

		// Prepend values with
		// the directive name.
		directiveValues := append(
			[]string{directive},
			values...,
		)

		// Space-separate them.
		policyDirective := strings.Join(directiveValues, " ")

		// Done.
		policyDirectives[i] = policyDirective
	}

	// Content-security-policy looks like this:
	// `Content-Security-Policy: <policy-directive>; <policy-directive>`
	// So join each policy directive appropriately.
	return strings.Join(policyDirectives, "; ")
}
