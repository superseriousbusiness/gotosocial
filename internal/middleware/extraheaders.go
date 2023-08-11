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
	"codeberg.org/gruf/go-debug"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// ExtraHeaders returns a new gin middleware which adds various extra headers to the response.
func ExtraHeaders() gin.HandlerFunc {
	csp := BuildContentSecurityPolicy()

	return func(c *gin.Context) {
		// Inform all callers which server implementation this is.
		c.Header("Server", "gotosocial")

		// Prevent google chrome cohort tracking. Originally this was referred
		// to as FlocBlock. Floc was replaced by Topics in 2022 and the spec says
		// that interest-cohort will also block Topics (as of 2022-Nov).
		//
		// See: https://smartframe.io/blog/google-topics-api-everything-you-need-to-know
		//
		// See: https://github.com/patcg-individual-drafts/topics
		c.Header("Permissions-Policy", "browsing-topics=()")

		// Inform the browser we only load
		// CSS/JS/media using the given policy.
		c.Header("Content-Security-Policy", csp)
	}
}

func BuildContentSecurityPolicy() string {
	// Start with restrictive policy.
	policy := "default-src 'self'"

	if debug.DEBUG {
		// Debug is enabled, allow
		// serving things from localhost
		// as well (regardless of port).
		policy += " localhost:*"
	}

	s3Endpoint := config.GetStorageS3Endpoint()
	if s3Endpoint == "" {
		// S3 not configured,
		// default policy is OK.
		return policy
	}

	if config.GetStorageS3Proxy() {
		// S3 is configured in proxy
		// mode, default policy is OK.
		return policy
	}

	// S3 is on and in non-proxy mode, so we need to add the S3 host to
	// the policy to allow images and video to be pulled from there too.

	// If secure is false,
	// use 'http' scheme.
	scheme := "https"
	if !config.GetStorageS3UseSSL() {
		scheme = "http"
	}

	// Construct endpoint URL.
	s3EndpointURLStr := scheme + "://" + s3Endpoint

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/img-src
	policy += "; img-src 'self' " + s3EndpointURLStr

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/media-src
	policy += "; media-src 'self' " + s3EndpointURLStr

	return policy
}
