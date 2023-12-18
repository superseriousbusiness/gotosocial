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

package headerfilter

import (
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"regexp"
)

// Maximum header value size before we return
// an instant negative match. They shouldn't
// go beyond this size in most cases anywho.
const MaxHeaderValue = 1024

// ErrLargeHeaderValue is returned on attempting to match on a value > MaxHeaderValue.
var ErrLargeHeaderValue = errors.New("header value too large")

// Filters represents a set of http.Header regular
// expression filters built-in statistic tracking.
type Filters []headerfilter

type headerfilter struct {
	// key is the header key to match against
	// in canonical textproto mime header format.
	key string

	// exprs contains regular expressions to
	// match values against for this header key.
	exprs []*regexp.Regexp
}

// Append will add new header filter expression under given header key.
func (fs *Filters) Append(key string, expr string) error {
	var filter *headerfilter

	// Ensure in canonical mime header format.
	key = textproto.CanonicalMIMEHeaderKey(key)

	// Look for existing filter
	// with key in filter slice.
	for i := range *fs {
		if (*fs)[i].key == key {
			filter = &((*fs)[i])
			break
		}
	}

	if filter == nil {
		// No existing filter found, create new.

		// Append new header filter to slice.
		(*fs) = append((*fs), headerfilter{})

		// Then take ptr to this new filter
		// at the last index in the slice.
		filter = &((*fs)[len((*fs))-1])

		// Setup new key.
		filter.key = key
	}

	// Compile regular expression.
	reg, err := regexp.Compile(expr)
	if err != nil {
		return fmt.Errorf("error compiling regexp %q: %w", expr, err)
	}

	// Append regular expression to filter.
	filter.exprs = append(filter.exprs, reg)

	return nil
}

// RegularMatch returns whether any values in http header
// matches any of the receiving filter regular expressions.
// This returns the matched header key, and matching regexp.
func (fs Filters) RegularMatch(h http.Header) (string, string, error) {
	for _, filter := range fs {
		for _, value := range h[filter.key] {
			// Don't perform match on large values
			// to mitigate denial of service attacks.
			if len(value) > MaxHeaderValue {
				return "", "", ErrLargeHeaderValue
			}

			// Compare against regular exprs.
			for _, expr := range filter.exprs {
				if expr.MatchString(value) {
					return filter.key, expr.String(), nil
				}
			}
		}
	}
	return "", "", nil
}

// InverseMatch returns whether any values in http header do
// NOT match any of the receiving filter regular expressions.
// This returns the matched header key, and matching regexp.
func (fs Filters) InverseMatch(h http.Header) (string, string, error) {
	for _, filter := range fs {
		for _, value := range h[filter.key] {
			// Don't perform match on large values
			// to mitigate denial of service attacks.
			if len(value) > MaxHeaderValue {
				return "", "", ErrLargeHeaderValue
			}

			// Compare against regular exprs.
			for _, expr := range filter.exprs {
				if !expr.MatchString(value) {
					return filter.key, expr.String(), nil
				}
			}
		}
	}
	return "", "", nil
}
