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

package gtscontext

import (
	"context"
	"time"
)

// WithValues wraps 'ctx' to use its deadline, done channel and error, but use value store of 'values'.
func WithValues(ctx context.Context, values context.Context) context.Context {
	if ctx == nil {
		panic("nil base context")
	}
	return &wrapContext{
		base: ctx,
		vals: values,
	}
}

type wrapContext struct {
	base context.Context
	vals context.Context
}

func (ctx *wrapContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.base.Deadline()
}

func (ctx *wrapContext) Done() <-chan struct{} {
	return ctx.base.Done()
}

func (ctx *wrapContext) Err() error {
	return ctx.base.Err()
}

func (ctx *wrapContext) Value(key any) any {
	return ctx.vals.Value(key)
}
