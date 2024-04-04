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

package httpclient

import (
	"net/http"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const (
	// starting backoff duration.
	baseBackoff = 2 * time.Second
)

// request wraps an HTTP request
// to add our own retry / backoff.
type request struct {

	// underlying request.
	req *http.Request

	// current backoff dur.
	backoff time.Duration

	// delivery attempts.
	attempts uint
}

// wrapRequest wraps an http.Request{} in our own request{} type.
func wrapRequest(req *http.Request) request {
	var r request
	r.req = req
	return r
}

// requestLog returns a prepared log entry with fields for http.Request{}.
func requestLog(r *http.Request) log.Entry {
	return log.WithContext(r.Context()).
		WithField("method", r.Method).
		WithField("url", r.URL.String())
}

// BackOff returns the currently set backoff duration,
// setting a default according to no. attempts if needed.
func (r *request) BackOff() time.Duration {
	if r.backoff <= 0 {
		// No backoff dur found, set our predefined
		// backoff according to a multiplier of 2^n.
		r.backoff = baseBackoff * 1 << (r.attempts + 1)
	}
	return r.backoff
}
