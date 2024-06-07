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

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
)

// SignFunc is a function signature that provides request signing.
type SignFunc func(r *http.Request) error

// signingtransport wraps an http.Transport{}
// (RoundTripper implementer) to check request
// context for a signing function and using for
// all subsequent trips through RoundTrip().
type signingtransport struct{ http.Transport }

func (t *signingtransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Ensure updated host always set.
	r.Header.Set("Host", r.URL.Host)

	if sign := gtscontext.HTTPClientSignFunc(r.Context()); sign != nil {
		// Reset signing header fields
		now := time.Now().UTC()
		r.Header.Set("Date", now.Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
		r.Header.Del("Signature")
		r.Header.Del("Digest")

		// Sign the outgoing request.
		if err := sign(r); err != nil {
			return nil, err
		}
	}

	// Pass to underlying transport.
	return t.Transport.RoundTrip(r)
}
