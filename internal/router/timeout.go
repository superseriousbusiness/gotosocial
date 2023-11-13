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

package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const requestTimeout = 10 * time.Minute

var ErrRequestTimeout = fmt.Errorf("timeoutHandler: timed out incoming HTTP request after %.0f minutes", requestTimeout.Minutes())

type timeoutHandler struct {
	*gin.Engine
}

// ServeHTTP wraps the embedded Gin engine's ServeHTTP
// function with an injected context which times out
// non-upgraded inbound requests after 10 minutes.
func (th timeoutHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	if upgr := r.Header.Get("Upgrade"); upgr != "" {
		// Upgrade to wss (probably).
		// Leave well enough alone.
		th.Engine.ServeHTTP(w, r)
		return
	}

	// Create timeout ctx.
	toCtx, cancelCtx := context.WithTimeoutCause(
		r.Context(),
		requestTimeout,
		ErrRequestTimeout,
	)
	defer cancelCtx()

	// Serve the request using a shallow copy
	// with the new context, without replacing
	// the underlying request, since the latter
	// may be used later outside of the Gin
	// engine for post-request cleanup tasks.
	th.Engine.ServeHTTP(w, r.WithContext(toCtx))
}
