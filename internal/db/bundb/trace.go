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

package bundb

import (
	"context"
	"database/sql"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func newDebugQueryHook() bun.QueryHook {
	return &debugQueryHook{}
}

// debugQueryHook implements bun.QueryHook
type debugQueryHook struct {
}

func (q *debugQueryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	// do nothing
	return ctx
}

// AfterQuery logs the time taken to query, the operation (select, update, etc), and the query itself as translated by bun.
func (q *debugQueryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	dur := time.Since(event.StartTime).Round(time.Microsecond)
	l := logrus.WithFields(logrus.Fields{
		"duration":  dur,
		"operation": event.Operation(),
	})

	if event.Err != nil && event.Err != sql.ErrNoRows {
		// if there's an error the it'll be handled in the application logic,
		// but we can still debug log it here alongside the query
		l = l.WithField("query", event.Query)
		l.Debug(event.Err)
		return
	}

	l.Tracef("[%s] %s", dur, event.Operation())
}
