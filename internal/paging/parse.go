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

package paging

import (
	"strconv"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// ParseIDPage parses an ID Page from a request context, returning BadRequest on error parsing.
// The min, max and default parameters define the page size limit minimum, maximum and default
// value, where a non-zero default will enforce paging for the endpoint on which this is called.
// While conversely, a zero default limit will not enforce paging, returning a nil page value.
func ParseIDPage(c *gin.Context, min, max, _default int) (*Page, gtserror.WithCode) {
	// Extract request query params.
	sinceID, haveSince := c.GetQuery("since_id")
	minID, haveMin := c.GetQuery("min_id")
	maxID, haveMax := c.GetQuery("max_id")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, min, max, _default)
	if errWithCode != nil {
		return nil, errWithCode
	}

	switch {
	case haveMin:
		// A min_id was supplied, even if the value
		// itself is empty. This indicates ASC order.
		return &Page{
			Min:   MinID(minID),
			Max:   MaxID(maxID),
			Limit: limit,
		}, nil

	case haveMax || haveSince:
		// A max_id or since_id was supplied, even if the
		// value itself is empty. This indicates DESC order.
		return &Page{
			Min:   SinceID(sinceID),
			Max:   MaxID(maxID),
			Limit: limit,
		}, nil

	case limit == 0:
		// No ID paging params provided, and no default
		// limit value which indicates paging not enforced.
		return nil, nil

	default:
		// only limit.
		return &Page{
			Min:   SinceID(""),
			Max:   MaxID(""),
			Limit: limit,
		}, nil
	}
}

// ParseShortcodeDomainPage parses an emoji shortcode domain Page from a request context, returning BadRequest
// on error parsing. The min, max and default parameters define the page size limit minimum, maximum and default
// value where a non-zero default will enforce paging for the endpoint on which this is called. While conversely,
// a zero default limit will not enforce paging, returning a nil page value.
func ParseShortcodeDomainPage(c *gin.Context, min, max, _default int) (*Page, gtserror.WithCode) {
	// Extract request query parameters.
	minShortcode, haveMin := c.GetQuery("min_shortcode_domain")
	maxShortcode, haveMax := c.GetQuery("max_shortcode_domain")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, min, max, _default)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !haveMin &&
		!haveMax &&
		limit == 0 {
		// No ID paging params provided, and no default
		// limit value which indicates paging not enforced.
		return nil, nil
	}

	return &Page{
		Min:   MinShortcodeDomain(minShortcode),
		Max:   MaxShortcodeDomain(maxShortcode),
		Limit: limit,
	}, nil
}

// ParseLimit parses the limit query parameter from a request context, returning BadRequest on error parsing and _default if zero limit given.
func ParseLimit(c *gin.Context, min, max, _default int) (int, gtserror.WithCode) {
	// Get limit query param.
	str, ok := c.GetQuery("limit")
	if !ok {
		return _default, nil
	}

	// Attempt to parse limit int.
	i, err := strconv.Atoi(str)
	if err != nil {
		const help = "bad integer limit value"
		return 0, gtserror.NewErrorBadRequest(err, help)
	}

	switch {
	case i == 0:
		return _default, nil
	case i < min:
		return min, nil
	case i > max:
		return max, nil
	default:
		return i, nil
	}
}
