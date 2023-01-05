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

package util

import (
	"fmt"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// PageableResponseParams models the parameters to pass to PackagePageableResponse.
//
// The given items will be provided in the paged response.
//
// The other values are all used to create the Link header so that callers know
// which endpoint to query next and previously in order to do paging.
type PageableResponseParams struct {
	Items            []interface{} // Sorted slice of items (statuses, notifications, etc)
	Path             string        // path to use for next/prev queries in the link header
	NextMaxIDKey     string        // key to use for the next max id query param in the link header, defaults to 'max_id'
	NextMaxIDValue   string        // value to use for next max id
	PrevMinIDKey     string        // key to use for the prev min id query param in the link header, defaults to 'min_id'
	PrevMinIDValue   string        // value to use for prev min id
	Limit            int           // limit number of entries to return
	ExtraQueryParams []string      // any extra query parameters to provide in the link header, should be in the format 'example=value'
}

// PackagePageableResponse is a convenience function for returning
// a bunch of pageable items (notifications, statuses, etc), as well
// as a Link header to inform callers of where to find next/prev items.
func PackagePageableResponse(params PageableResponseParams) (*apimodel.PageableResponse, gtserror.WithCode) {
	if params.NextMaxIDKey == "" {
		params.NextMaxIDKey = "max_id"
	}

	if params.PrevMinIDKey == "" {
		params.PrevMinIDKey = "min_id"
	}

	pageableResponse := EmptyPageableResponse()

	if len(params.Items) == 0 {
		return pageableResponse, nil
	}

	// items
	pageableResponse.Items = params.Items

	protocol := config.GetProtocol()
	host := config.GetHost()

	// next
	nextRaw := params.NextMaxIDKey + "=" + params.NextMaxIDValue
	if params.Limit != 0 {
		nextRaw = fmt.Sprintf("limit=%d&", params.Limit) + nextRaw
	}
	for _, p := range params.ExtraQueryParams {
		nextRaw = nextRaw + "&" + p
	}
	nextLink := &url.URL{
		Scheme:   protocol,
		Host:     host,
		Path:     params.Path,
		RawQuery: nextRaw,
	}
	nextLinkString := nextLink.String()
	pageableResponse.NextLink = nextLinkString

	// prev
	prevRaw := params.PrevMinIDKey + "=" + params.PrevMinIDValue
	if params.Limit != 0 {
		prevRaw = fmt.Sprintf("limit=%d&", params.Limit) + prevRaw
	}
	for _, p := range params.ExtraQueryParams {
		prevRaw = prevRaw + "&" + p
	}
	prevLink := &url.URL{
		Scheme:   protocol,
		Host:     host,
		Path:     params.Path,
		RawQuery: prevRaw,
	}
	prevLinkString := prevLink.String()
	pageableResponse.PrevLink = prevLinkString

	// link header
	next := fmt.Sprintf("<%s>; rel=\"next\"", nextLinkString)
	prev := fmt.Sprintf("<%s>; rel=\"prev\"", prevLinkString)
	pageableResponse.LinkHeader = next + ", " + prev

	return pageableResponse, nil
}

// EmptyPageableResponse just returns an empty
// PageableResponse with no link header or items.
func EmptyPageableResponse() *apimodel.PageableResponse {
	return &apimodel.PageableResponse{
		Items: []interface{}{},
	}
}
