/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package search

import (
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// BasePathV1 is the base path for serving v1 of the search API
	BasePathV1 = "/api/v1/search"

	// BasePathV2 is the base path for serving v2 of the search API
	BasePathV2 = "/api/v2/search"

	// AccountIDKey -- If provided, statuses returned will be authored only by this account
	AccountIDKey = "account_id"
	// MaxIDKey -- Return results older than this id
	MaxIDKey = "max_id"
	// MinIDKey -- Return results immediately newer than this id
	MinIDKey = "min_id"
	// TypeKey -- Enum(accounts, hashtags, statuses)
	TypeKey = "type"
	// ExcludeUnreviewedKey -- Filter out unreviewed tags? Defaults to false. Use true when trying to find trending tags.
	ExcludeUnreviewedKey = "exclude_unreviewed"
	// QueryKey -- The search query
	QueryKey = "q"
	// ResolveKey -- Attempt WebFinger lookup. Defaults to false.
	ResolveKey = "resolve"
	// LimitKey -- Maximum number of results to load, per type. Defaults to 20. Max 40.
	LimitKey = "limit"
	// OffsetKey -- Offset in search results. Used for pagination. Defaults to 0.
	OffsetKey = "offset"
	// FollowingKey -- Only include accounts that the user is following. Defaults to false.
	FollowingKey = "following"

	// TypeAccounts --
	TypeAccounts = "accounts"
	// TypeHashtags --
	TypeHashtags = "hashtags"
	// TypeStatuses --
	TypeStatuses = "statuses"
)

// Module implements the ClientAPIModule interface for everything related to searching
type Module struct {
	config    *config.Config
	processor processing.Processor
}

// New returns a new search module
func New(config *config.Config, processor processing.Processor) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodGet, BasePathV1, m.SearchGETHandler)
	r.AttachHandler(http.MethodGet, BasePathV2, m.SearchGETHandler)
	return nil
}
