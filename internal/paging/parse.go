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

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// ParseIDPage parses an ID Page from a request context, returning BadRequest on error parsing.
// The min, max and default parameters define the page size limit minimum, maximum and default
// value, where a non-zero default will enforce paging for the endpoint on which this is called.
// While conversely, a zero default limit will not enforce paging, returning a nil page value.
func ParseIDPage(c *gin.Context, min, max, _default int) (*Page, gtserror.WithCode) {
	// Extract request query params.
	sinceID := c.Query("since_id")
	minID := c.Query("min_id")
	maxID := c.Query("max_id")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, min, max, _default)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if sinceID == "" &&
		minID == "" &&
		maxID == "" &&
		limit == 0 {
		// No ID paging params provided, and no default
		// limit value which indicates paging not enforced.
		return nil, nil
	}

	return &Page{
		Min:   MinID(minID, sinceID),
		Max:   MaxID(maxID),
		Limit: limit,
	}, nil
}

// ParseShortcodeDomainPage parses an emoji shortcode domain Page from a request context, returning BadRequest
// on error parsing. The min, max and default parameters define the page size limit minimum, maximum and default
// value where a non-zero default will enforce paging for the endpoint on which this is called. While conversely,
// a zero default limit will not enforce paging, returning a nil page value.
func ParseShortcodeDomainPage(c *gin.Context, min, max, _default int) (*Page, gtserror.WithCode) {
	// Extract request query parameters.
	minShortcode := c.Query("min_shortcode_domain")
	maxShortcode := c.Query("max_shortcode_domain")

	// Extract request limit parameter.
	limit, errWithCode := ParseLimit(c, min, max, _default)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if minShortcode == "" &&
		maxShortcode == "" &&
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
	str := c.Query("limit")
	if str == "" {
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
