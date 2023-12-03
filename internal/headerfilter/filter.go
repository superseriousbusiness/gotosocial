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
	"fmt"
	"net/http"
	"net/textproto"
	"regexp"
	"slices"
	"sync/atomic"
)

// Filters represents a set of http.Header regular
// expression filters built-in statistic tracking.
type Filters []headerfilter

type headerfilter struct {
	// key is the header key to match against
	// in canonical textproto mime header format.
	key string

	// exprs contains regular expressions to
	// match values against for this header key.
	exprs []regexpr
}

// regexpr wraps a regular expression
// with an atomically updated uint64 in
// order to count no. positive matches.
type regexpr struct {
	*regexp.Regexp

	// match count.
	n atomic.Uint64
}

// Append will add new header filter expression under given header key.
func (fs *Filters) Append(key string, expr string) error {
	var filter *headerfilter

	// Get key in canonical mime header format.
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
		(*fs) = append((*fs), headerfilter{})
		filter = &((*fs)[len((*fs))-1])
		filter.key = key
	}

	// Compile regular expression.
	regExpr, err := regexp.Compile(expr)
	if err != nil {
		return fmt.Errorf("error compiling regexp %q: %w", expr, err)
	}

	// Append wrapped expression type to filter.
	filter.exprs = append(filter.exprs, regexpr{
		Regexp: regExpr,
	})

	return nil
}

// MatchPositive
func (fs Filters) MatchPositive(h http.Header) bool {
	for _, filter := range fs {
		for _, value := range h[filter.key] {
			// Shorten header value if needed
			// to mitigate denial of service.
			value = safeHeaderValue(value)

			// Compare against regexprs.
			for i := range filter.exprs {
				if filter.exprs[i].MatchString(value) {
					filter.exprs[i].n.Add(1)
					return false
				}
			}
		}
	}
	return true
}

// MatchNegative
func (fs Filters) MatchNegative(h http.Header) bool {
	for _, filter := range fs {
		for _, value := range h[filter.key] {
			// Shorten header value if needed
			// to mitigate denial of service.
			value = safeHeaderValue(value)

			// Compare against regexprs.
			for i := range filter.exprs {
				if filter.exprs[i].MatchString(value) {
					filter.exprs[i].n.Add(1)
					return true
				}
			}
		}
	}
	return false
}

// Stats compiles each of the filters and their match counts into
// a readable set of HeaderKey:{ValueExpr: MatchCount} statistics.
// TODO: may be worth updating this to be more prometheus readable
func (fs Filters) Stats() map[string]map[string]uint64 {
	stats := make(map[string]map[string]uint64, len(fs))
	for _, filter := range fs {
		// Allocate and append map for this filter's stats.
		fstats := make(map[string]uint64, len(filter.exprs))
		stats[filter.key] = fstats

		// Append all expression stats.
		for i := range filter.exprs {
			e := &(filter.exprs[i])
			fstats[e.String()] = e.n.Load()
		}
	}
	return stats
}

// safeHeaderValue returns a header value checked against
// a constant max length, else returning a subset of value
// amounting to header[:max/2] + header[len(header)+max/2:].
func safeHeaderValue(value string) string {
	const max = 1024
	const mid = max / 2
	if len(value) <= max {
		return value
	}
	return value[:mid] + value[len(value)-1-mid:]
}

// dedupeExprs removes duplicates from a passed
// slice of string regular expressions using a map.
func deduplicateExprs(exprs []string) []string {
	dedupe := make(map[string]struct{}, len(exprs))
	return slices.DeleteFunc(exprs, func(expr string) bool {
		if _, ok := dedupe[expr]; ok {
			return true
		}
		dedupe[expr] = struct{}{}
		return false
	})
}
