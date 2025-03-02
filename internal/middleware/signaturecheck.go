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
	"context"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/log"

	"codeberg.org/superseriousbusiness/httpsig"
	"github.com/gin-gonic/gin"
)

const (
	sigHeader  = string(httpsig.Signature)
	authHeader = string(httpsig.Authorization)
	// untyped error returned by httpsig when no signature is present
	noSigError = "neither \"" + sigHeader + "\" nor \"" + authHeader + "\" have signature parameters"
)

// SignatureCheck returns a gin middleware for checking http signatures.
//
// The middleware first checks whether an incoming http request has been
// http-signed with a well-formed signature. If so, it will check if the
// domain that signed the request is permitted to access the server, using
// the provided uriBlocked function. If the domain is blocked, the middleware
// will abort the request chain with http code 403 forbidden. If it is not
// blocked, the handler will set the key verifier and the signature in the
// context for use down the line.
//
// In case of an error, the request will be aborted with http code 500.
func SignatureCheck(uriBlocked func(context.Context, *url.URL) (bool, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Create the signature verifier from the request;
		// this will error if the request wasn't signed.
		verifier, err := httpsig.NewVerifier(c.Request)
		if err != nil {
			// Only actually *abort* the request with 401
			// if a signature was present but malformed.
			// Otherwise proceed with an unsigned request;
			// it's up to other functions to reject this.
			if err.Error() != noSigError {
				log.Debugf(ctx, "http signature was present but invalid: %s", err)
				c.AbortWithStatus(http.StatusUnauthorized)
			}

			return
		}

		// The request was signed! The key ID should be given
		// in the signature so that we know where to fetch it
		// from the remote server. This will be something like:
		// https://example.org/users/some_remote_user#main-key
		pubKeyIDStr := verifier.KeyId()

		// Key can sometimes be nil, according to url parse
		// func: 'Trying to parse a hostname and path without
		// a scheme is invalid but may not necessarily return
		// an error, due to parsing ambiguities'. Catch this.
		pubKeyID, err := url.Parse(pubKeyIDStr)
		if err != nil || pubKeyID == nil {
			log.Warnf(ctx, "pubkey id %s could not be parsed as a url", pubKeyIDStr)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// If the domain is blocked we want to bail as fast as
		// possible without the request proceeding further.
		blocked, err := uriBlocked(ctx, pubKeyID)
		if err != nil {
			log.Errorf(ctx, "error checking block for domain %s: %s", pubKeyID.Host, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if blocked {
			log.Infof(ctx, "domain %s is blocked", pubKeyID.Host)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Assume signature was set on Signature header,
		// but fall back to Authorization header if necessary.
		signature := c.GetHeader(sigHeader)
		if signature == "" {
			signature = c.GetHeader(authHeader)
		}

		// Set relevant values on the request context
		// to save some work further down the line.
		ctx = gtscontext.SetHTTPSignatureVerifier(ctx, verifier)
		ctx = gtscontext.SetHTTPSignature(ctx, signature)
		ctx = gtscontext.SetHTTPSignaturePubKeyID(ctx, pubKeyID)

		// Replace request with a shallow
		// copy with the new context.
		c.Request = c.Request.WithContext(ctx)
	}
}
