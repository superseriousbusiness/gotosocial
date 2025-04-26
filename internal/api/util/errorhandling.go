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
	"context"
	"errors"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-kv"
	"github.com/gin-gonic/gin"
)

// TODO: add more templated html pages here for different error types

// NotFoundHandler serves a 404 html page through the provided gin context,
// if accept is 'text/html', or just returns a json error if 'accept' is empty
// or application/json.
//
// When serving html, NotFoundHandler calls the provided InstanceGet function
// to fetch the apimodel representation of the instance, for serving in the
// 404 header and footer.
//
// If an error is returned by InstanceGet, the function will panic.
func NotFoundHandler(c *gin.Context, instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode), accept string, errWithCode gtserror.WithCode) {
	switch accept {
	case string(TextHTML):
		ctx := c.Request.Context()
		instance, err := instanceGet(ctx)
		if err != nil {
			panic(err)
		}

		template404Page(c,
			instance,
			gtscontext.RequestID(ctx),
		)
	default:
		JSON(c, http.StatusNotFound, map[string]string{
			"error": errWithCode.Safe(),
		})
	}
}

// genericErrorHandler is a more general version of the NotFoundHandler, which can
// be used for serving either generic error pages with some rendered help text,
// or just some error json if the caller prefers (or has no preference).
func genericErrorHandler(c *gin.Context, instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode), accept string, errWithCode gtserror.WithCode) {
	switch accept {
	case string(TextHTML):
		ctx := c.Request.Context()
		instance, err := instanceGet(ctx)
		if err != nil {
			panic(err)
		}

		templateErrorPage(c,
			instance,
			errWithCode.Code(),
			errWithCode.Safe(),
			gtscontext.RequestID(ctx),
		)
	default:
		JSON(c, errWithCode.Code(), map[string]string{
			"error": errWithCode.Safe(),
		})
	}
}

// ErrorHandler takes the provided gin context and errWithCode
// and tries to serve a helpful error to the caller.
//
// It will do content negotiation to figure out if the caller prefers
// to see an html page with the error rendered there. If not, or if
// something goes wrong during the function, it will recover and just
// try to serve an appropriate application/json content-type error.
// To override the default response type, specify `offers`.
//
// If the requester already hung up on the request, or the server
// timed out a very slow request, ErrorHandler will overwrite the
// given errWithCode with a 408 or 499 error to indicate that the
// failure wasn't due to something we did, and will avoid trying
// to write extensive bytes to the caller by just aborting.
//
// For 499, see https://en.wikipedia.org/wiki/List_of_HTTP_status_codes#nginx.
func ErrorHandler(
	c *gin.Context,
	errWithCode gtserror.WithCode,
	instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode),
	offers ...string,
) {
	if ctxErr := c.Request.Context().Err(); ctxErr != nil {
		// Context error means either client has left already,
		// or server has timed out a very slow request.
		//
		// Rewrap the error with something less scary,
		// and just abort the request gracelessly.
		err := errWithCode.Unwrap()

		if errors.Is(ctxErr, context.DeadlineExceeded) {
			// We timed out the request.
			errWithCode = gtserror.NewErrorRequestTimeout(err)

			// Be correct and write "close".
			// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Connection#close
			// and: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/408
			c.Header("Connection", "close")
		} else {
			// Client timed out the request.
			errWithCode = gtserror.NewErrorClientClosedRequest(err)
		}

		c.AbortWithStatus(errWithCode.Code())
		return
	}

	// Set the error on the gin context so that it can be logged
	// in the gin logger middleware (internal/middleware/logger.go).
	c.Error(errWithCode) //nolint:errcheck

	// Discover if we're allowed to serve a nice html error page,
	// or if we should just use a json. Normally we would want to
	// check for a returned error, but if an error occurs here we
	// can just fall back to default behavior (serve json error).
	// Prefer provided offers, fall back to JSON or HTML.
	accept, _ := NegotiateAccept(c, append(offers, JSONOrHTMLAcceptHeaders...)...)

	if errWithCode.Code() == http.StatusNotFound {
		// Use our special not found handler with useful status text.
		NotFoundHandler(c, instanceGet, accept, errWithCode)
	} else {
		genericErrorHandler(c, instanceGet, accept, errWithCode)
	}
}

// WebErrorHandler is like ErrorHandler, but will display HTML over JSON by default.
func WebErrorHandler(c *gin.Context, errWithCode gtserror.WithCode, instanceGet func(ctx context.Context) (*apimodel.InstanceV1, gtserror.WithCode)) {
	ErrorHandler(c, errWithCode, instanceGet, TextHTML, AppJSON)
}

// OAuthErrorHandler is a lot like ErrorHandler, but it specifically returns errors
// that are compatible with https://datatracker.ietf.org/doc/html/rfc6749#section-5.2,
// but serializing errWithCode.Error() in the 'error' field, and putting any help text
// from the error in the 'error_description' field. This means you should be careful not
// to pass any detailed errors (that might contain sensitive information) into the
// errWithCode.Error() field, since the client will see this. Use your noggin!
func OAuthErrorHandler(c *gin.Context, errWithCode gtserror.WithCode) {
	l := log.WithContext(c.Request.Context()).
		WithFields(kv.Fields{
			{"path", c.Request.URL.Path},
			{"error", errWithCode.Error()},
			{"help", errWithCode.Safe()},
		}...)

	statusCode := errWithCode.Code()

	if statusCode == http.StatusInternalServerError {
		l.Error("Internal Server Error")
	} else {
		l.Debug("handling OAuth error")
	}

	JSON(c, statusCode, map[string]string{
		"error":             errWithCode.Error(),
		"error_description": errWithCode.Safe(),
	})
}

// NotFoundAfterMove returns code 404 to the caller and writes a helpful error message.
// Specifically used for accounts trying to access endpoints they cannot use while moving.
func NotFoundAfterMove(c *gin.Context) {
	const errMsg = "your account has Moved or is currently Moving; you cannot use this endpoint"
	JSON(c, http.StatusForbidden, map[string]string{
		"error": errMsg,
	})
}

// ForbiddenAfterMove returns code 403 to the caller and writes a helpful error message.
// Specifically used for accounts trying to take actions on endpoints they cannot do while moving.
func ForbiddenAfterMove(c *gin.Context) {
	const errMsg = "your account has Moved or is currently Moving; you cannot take create or update type actions"
	JSON(c, http.StatusForbidden, map[string]string{
		"error": errMsg,
	})
}
