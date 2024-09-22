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
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/headerfilter"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

var (
	allowMatches = matchstats{m: make(map[string]uint64)}
	blockMatches = matchstats{m: make(map[string]uint64)}

	errHeaderNotAllowed = errors.New("header did not match allow filter")
	errHeaderBlocked    = errors.New("header matched block filter")
)

// matchstats is a simple statistics
// counter for header filter matches.
// TODO: replace with otel.
type matchstats struct {
	m map[string]uint64
	l sync.Mutex
}

func (m *matchstats) Add(hdr, regex string) {
	m.l.Lock()
	key := hdr + ":" + regex
	m.m[key]++
	m.l.Unlock()
}

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
	_ = *state //nolint
	// Allowlist mode: explicit block takes
	// precedence over explicit allow.
	//
	// Headers that have neither block
	// or allow entries are blocked.
	return func(c *gin.Context) {

		// Check if header is explicitly blocked.
		block, err := isHeaderBlocked(state, c)
		if err != nil {
			respondInternalServerError(c, err)
			return
		}

		if block {
			c.Error(errHeaderBlocked)
			respondBlocked(c)
			return
		}

		// Check if header is missing explicit allow.
		notAllow, err := isHeaderNotAllowed(state, c)
		if err != nil {
			respondInternalServerError(c, err)
			return
		}

		if notAllow {
			c.Error(errHeaderNotAllowed)
			respondBlocked(c)
			return
		}

		// Allowed!
		c.Next()
	}
}

func headerFilterBlockMode(state *state.State) func(c *gin.Context) {
	_ = *state //nolint
	// Blocklist/default mode: explicit allow
	// takes precedence over explicit block.
	//
	// Headers that have neither block
	// or allow entries are allowed.
	return func(c *gin.Context) {

		// Check if header is explicitly allowed.
		allow, err := isHeaderAllowed(state, c)
		if err != nil {
			respondInternalServerError(c, err)
			return
		}

		if !allow {
			// Check if header is explicitly blocked.
			block, err := isHeaderBlocked(state, c)
			if err != nil {
				respondInternalServerError(c, err)
				return
			}

			if block {
				c.Error(errHeaderBlocked)
				respondBlocked(c)
				return
			}
		}

		// Allowed!
		c.Next()
	}
}

func isHeaderBlocked(state *state.State, c *gin.Context) (bool, error) {
	var (
		ctx = c.Request.Context()
		hdr = c.Request.Header
	)

	// Perform an explicit is-blocked check on request header.
	key, expr, err := state.DB.BlockHeaderRegularMatch(ctx, hdr)
	switch err {
	case nil:
		break

	case headerfilter.ErrLargeHeaderValue:
		log.Warn(ctx, "large header value")
		key = "*" // block large headers

	default:
		err := gtserror.Newf("error checking header: %w", err)
		return false, err
	}

	if key != "" {
		if expr != "" {
			// Increment block matches stat.
			// TODO: replace expvar with build
			// taggable metrics types in State{}.
			blockMatches.Add(key, expr)
		}

		// A header was matched against!
		// i.e. this request is blocked.
		return true, nil
	}

	return false, nil
}

func isHeaderAllowed(state *state.State, c *gin.Context) (bool, error) {
	var (
		ctx = c.Request.Context()
		hdr = c.Request.Header
	)

	// Perform an explicit is-allowed check on request header.
	key, expr, err := state.DB.AllowHeaderRegularMatch(ctx, hdr)
	switch err {
	case nil:
		break

	case headerfilter.ErrLargeHeaderValue:
		log.Warn(ctx, "large header value")
		key = "" // block large headers

	default:
		err := gtserror.Newf("error checking header: %w", err)
		return false, err
	}

	if key != "" {
		if expr != "" {
			// Increment allow matches stat.
			// TODO: replace expvar with build
			// taggable metrics types in State{}.
			allowMatches.Add(key, expr)
		}

		// A header was matched against!
		// i.e. this request is allowed.
		return true, nil
	}

	return false, nil
}

func isHeaderNotAllowed(state *state.State, c *gin.Context) (bool, error) {
	var (
		ctx = c.Request.Context()
		hdr = c.Request.Header
	)

	// Perform an explicit is-NOT-allowed check on request header.
	key, expr, err := state.DB.AllowHeaderInverseMatch(ctx, hdr)
	switch err {
	case nil:
		break

	case headerfilter.ErrLargeHeaderValue:
		log.Warn(ctx, "large header value")
		key = "*" // block large headers

	default:
		err := gtserror.Newf("error checking header: %w", err)
		return false, err
	}

	if key != "" {
		if expr != "" {
			// Increment allow matches stat.
			// TODO: replace expvar with build
			// taggable metrics types in State{}.
			allowMatches.Add(key, expr)
		}

		// A header was matched against!
		// i.e. request is NOT allowed.
		return true, nil
	}

	return false, nil
}
