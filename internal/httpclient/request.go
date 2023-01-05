/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package httpclient

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// ValidateRequest performs the same request validation logic found in the default
// net/http.Transport{}.roundTrip() function, but pulls it out into this separate
// function allowing validation errors to be wrapped under a single error type.
func ValidateRequest(r *http.Request) error {
	switch {
	case r.URL == nil:
		return fmt.Errorf("%w: nil url", ErrInvalidRequest)
	case r.Header == nil:
		return fmt.Errorf("%w: nil header", ErrInvalidRequest)
	case r.URL.Host == "":
		return fmt.Errorf("%w: empty url host", ErrInvalidRequest)
	case r.URL.Scheme != "http" && r.URL.Scheme != "https":
		return fmt.Errorf("%w: unsupported protocol %q", ErrInvalidRequest, r.URL.Scheme)
	case strings.IndexFunc(r.Method, func(r rune) bool { return !httpguts.IsTokenRune(r) }) != -1:
		return fmt.Errorf("%w: invalid method %q", ErrInvalidRequest, r.Method)
	}

	for key, values := range r.Header {
		// Check field key name is valid
		if !httpguts.ValidHeaderFieldName(key) {
			return fmt.Errorf("%w: invalid header field name %q", ErrInvalidRequest, key)
		}

		// Check each field value is valid
		for i := 0; i < len(values); i++ {
			if !httpguts.ValidHeaderFieldValue(values[i]) {
				return fmt.Errorf("%w: invalid header field value %q", ErrInvalidRequest, values[i])
			}
		}
	}

	// ps. kim wrote this

	return nil
}
