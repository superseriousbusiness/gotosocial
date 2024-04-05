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

// Request wraps an HTTP request
// to add our own retry / backoff.
type Request struct {
	// Current backoff dur.
	backoff time.Duration

	// Delivery attempts.
	attempts uint

	// done is marked when
	// no more requests may
	// be attempted.
	done bool

	// log fields.
	log.Entry

	// underlying request.
	*http.Request
}

// WrapRequest wraps an existing http.Request within
// our own httpclient.Request with retry / backoff tracking.
func WrapRequest(r *http.Request) Request {
	var rr Request
	rr.Request = r
	rr.Entry = log.WithContext(r.Context()).
		WithField("method", r.Method).
		WithField("url", r.URL.String()).
		WithField("contentType", r.Header.Get("Content-Type"))
	return rr
}

// GetBackOff returns the currently set backoff duration,
// (using a default according to no. attempts if needed).
func (r *Request) BackOff() time.Duration {
	if r.backoff <= 0 {
		// No backoff dur found, set our predefined
		// backoff according to a multiplier of 2^n.
		r.backoff = baseBackoff * 1 << (r.attempts + 1)
	}
	return r.backoff
}
