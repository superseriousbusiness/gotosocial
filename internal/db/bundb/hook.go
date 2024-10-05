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

package bundb

import (
	"context"
	"time"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

// queryHook implements bun.QueryHook
type queryHook struct{}

func (queryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx // do nothing
}

// AfterQuery logs the time taken to query, the operation (select, update, etc), and the query itself as translated by bun.
func (queryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	// Get the DB query duration
	dur := time.Since(event.StartTime)

	switch {
	// Warn on slow database queries
	case dur > time.Second:
		log.WithContext(ctx).
			WithFields(kv.Fields{
				{"duration", dur},
				{"query", event.Query},
			}...).Warn("SLOW DATABASE QUERY")

	// On trace, we log query information,
	// manually crafting so DB query not escaped.
	case log.Level() >= log.TRACE:
		log.Printf("level=TRACE duration=%s query=%s", dur, event.Query)
	}
}
