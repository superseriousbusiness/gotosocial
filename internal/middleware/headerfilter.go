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

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/headerfilter"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// HeaderFilter returns a gin middleware handler that provides HTTP
// request blocking (filtering) based on database allow / block filters.
func HeaderFilter(state *state.State) gin.HandlerFunc {
	switch mode := config.GetAdvancedHeaderFilterMode(); mode {
	case config.RequestHeaderFilterModeDisabled:
		return func(ctx *gin.Context) {}

	case config.RequestHeaderFilterModeAllow:
		return headerFilterAllowMode(state)

	case config.RequestHeaderFilterModeBlock:
		return headerFilterBlockMode(state)

	default:
		panic("unrecognized filter mode: " + mode)
	}
}

func headerFilterAllowMode(state *state.State) func(c *gin.Context) {
	// Allowlist mode: explicit block takes
	// precedence over explicit allow.
	//
	// Headers that have neither block
	// or allow entries are blocked.
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		hdr := c.Request.Header

		// Perform an explicit block match, this skips allow match.
		block, err := state.DB.BlockHeaderRegularMatch(ctx, hdr)
		switch err {
		case nil:

		case headerfilter.ErrLargeHeaderValue:
			log.Warn(ctx, "large header value")
			block = true // always block

		default:
			respondInternalServerError(c, err)
			return
		}

		if block {
			respondBlocked(c)
			return
		}

		// Headers not explicitly blocked, check for allow NON-matches.
		notAllow, err := state.DB.AllowHeaderInverseMatch(ctx, hdr)
		switch err {
		case nil:

		case headerfilter.ErrLargeHeaderValue:
			log.Warn(ctx, "large header value")
			notAllow = true // always block

		default:
			respondInternalServerError(c, err)
			return
		}

		if notAllow {
			respondBlocked(c)
			return
		}

		// Allowed!
		c.Next()
	}
}

func headerFilterBlockMode(state *state.State) func(c *gin.Context) {
	// Blocklist/default mode: explicit allow
	// takes precedence over explicit block.
	//
	// Headers that have neither block
	// or allow entries are allowed.
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		hdr := c.Request.Header

		// Perform an explicit allow match, this skips block match.
		allow, err := state.DB.AllowHeaderRegularMatch(ctx, hdr)
		switch err {
		case nil:

		case headerfilter.ErrLargeHeaderValue:
			log.Warn(ctx, "large header value")
			respondBlocked(c) // always block
			return

		default:
			respondInternalServerError(c, err)
			return
		}

		if !allow {
			// Headers were not explicitly allowed, perform block match.
			block, err := state.DB.BlockHeaderRegularMatch(ctx, hdr)
			switch err {
			case nil:

			case headerfilter.ErrLargeHeaderValue:
				log.Warn(ctx, "large header value")
				block = true // always block

			default:
				respondInternalServerError(c, err)
				return
			}

			if block {
				respondBlocked(c)
				return
			}
		}

		// Allowed!
		c.Next()
	}
}
