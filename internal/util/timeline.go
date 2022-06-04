/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
)

// TimelineableResponseParams models the parameters to pass to PackageTimelineableResponse.
//
// The given items will be provided in the timeline response.
//
// The other values are all used to create the Link header so that callers know
// which endpoint to query next and previously in order to do paging.
type TimelineableResponseParams struct {
	Items            []timeline.Timelineable // Sorted slice of Timelineables (statuses, notifications, etc)
	Path             string                  // path to use for next/prev queries in the link header
	NextMaxIDKey     string                  // key to use for the next max id query param in the link header, defaults to 'max_id'
	NextMaxIDValue   string                  // value to use for next max id
	PrevMinIDKey     string                  // key to use for the prev min id query param in the link header, defaults to 'min_id'
	PrevMinIDValue   string                  // value to use for prev min id
	Limit            int                     // limit number of entries to return
	ExtraQueryParams []string                // any extra query parameters to provide in the link header, should be in the format 'example=value'
}

// PackageTimelineableResponse is a convenience function for returning
// a bunch of timelineable items (notifications, statuses, etc), as well
// as a Link header to inform callers of where to find next/prev items.
func PackageTimelineableResponse(params TimelineableResponseParams) (*apimodel.TimelineResponse, gtserror.WithCode) {
	if params.NextMaxIDKey == "" {
		params.NextMaxIDKey = "max_id"
	}

	if params.PrevMinIDKey == "" {
		params.PrevMinIDKey = "min_id"
	}

	timelineResponse := &apimodel.TimelineResponse{
		Items: params.Items,
	}

	// prepare the next and previous links
	if len(params.Items) != 0 {
		protocol := config.GetProtocol()
		host := config.GetHost()

		nextRaw := fmt.Sprintf("limit=%d&%s=%s", params.Limit, params.NextMaxIDKey, params.NextMaxIDValue)
		for _, p := range params.ExtraQueryParams {
			nextRaw = nextRaw + "&" + p
		}
		nextLink := &url.URL{
			Scheme:   protocol,
			Host:     host,
			Path:     params.Path,
			RawQuery: nextRaw,
		}
		next := fmt.Sprintf("<%s>; rel=\"next\"", nextLink.String())

		prevRaw := fmt.Sprintf("limit=%d&%s=%s", params.Limit, params.PrevMinIDKey, params.PrevMinIDValue)
		for _, p := range params.ExtraQueryParams {
			prevRaw = prevRaw + "&" + p
		}
		prevLink := &url.URL{
			Scheme:   protocol,
			Host:     host,
			Path:     params.Path,
			RawQuery: prevRaw,
		}
		prev := fmt.Sprintf("<%s>; rel=\"prev\"", prevLink.String())
		timelineResponse.LinkHeader = next + ", " + prev
	}

	return timelineResponse, nil
}

// EmptyTimelineResponse just returns an empty
// TimelineResponse with no link header or items.
func EmptyTimelineResponse() *apimodel.TimelineResponse {
	return &apimodel.TimelineResponse{
		Items: []timeline.Timelineable{},
	}
}
