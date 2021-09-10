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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

// queryHook is just a wrapper for bun.QueryHook
type queryHook bun.QueryHook

// debugQueryHook implements queryHook
type debugQueryHook struct {
	log *logrus.Logger
}

func (q *debugQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	// do nothing
	return ctx
}

// AfterQuery logs the time taken to query, the operation (select, update, etc), and the query itself as translated by bun.
func (q *debugQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	dur := time.Now().Sub(event.StartTime).Round(time.Microsecond)
	l := q.log.WithFields(logrus.Fields{
		"queryTime": dur,
		"operation": event.Operation(),
	})
	l.Trace(event.Query)
}
