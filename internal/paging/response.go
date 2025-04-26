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
	"net/url"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
)

// ResponseParams models the parameters to pass to PageableResponse.
//
// The given items will be provided in the paged response.
//
// The other values are all used to create the Link header so that callers know
// which endpoint to query next and previously in order to do paging.
type ResponseParams struct {
	Items []interface{} // Sorted slice of items (statuses, notifications, etc)
	Path  string        // path to use for next/prev queries in the link header
	Next  *Page         // page details for the next page
	Prev  *Page         // page details for the previous page
	Query url.Values    // any extra query parameters to provide in the link header, should be in the format 'example=value'
}

// PackageResponse is a convenience function for returning
// a bunch of pageable items (notifications, statuses, etc), as well
// as a Link header to inform callers of where to find next/prev items.
func PackageResponse(params ResponseParams) *apimodel.PageableResponse {
	var (
		// Extract paging params.
		nextPg = params.Next
		prevPg = params.Prev

		// Host app configuration.
		proto = config.GetProtocol()
		host  = config.GetHost()

		// Combined next/prev page link header parts.
		linkHeaderParts = make([]string, 0, 2)
	)

	// Build the next / previous page links from page and host config.
	nextLink := nextPg.ToLink(proto, host, params.Path, params.Query)
	prevLink := prevPg.ToLink(proto, host, params.Path, params.Query)

	if nextLink != "" {
		// Append page "next" link to header parts.
		linkHeaderParts = append(linkHeaderParts, `<`+nextLink+`>; rel="next"`)
	}

	if prevLink != "" {
		// Append page "prev" link to header parts.
		linkHeaderParts = append(linkHeaderParts, `<`+prevLink+`>; rel="prev"`)
	}

	return &apimodel.PageableResponse{
		Items:      params.Items,
		NextLink:   nextLink,
		PrevLink:   prevLink,
		LinkHeader: strings.Join(linkHeaderParts, ", "),
	}
}

// EmptyResponse just returns an empty
// PageableResponse with no link header or items.
func EmptyResponse() *apimodel.PageableResponse {
	return &apimodel.PageableResponse{
		Items: []interface{}{},
	}
}
