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

package log

import (
	"context"
	"fmt"

	"codeberg.org/gruf/go-kv"
)

type Entry struct {
	ctx context.Context
	kvs []kv.Field
}

// WithContext updates Entry{} value context.
func (e Entry) WithContext(ctx context.Context) Entry {
	e.ctx = ctx
	return e
}

// WithField appends key-value field to Entry{}.
func (e Entry) WithField(key string, value interface{}) Entry {
	e.kvs = append(e.kvs, kv.Field{K: key, V: value})
	return e
}

// WithFields appends key-value fields to Entry{}.
func (e Entry) WithFields(kvs ...kv.Field) Entry {
	e.kvs = append(e.kvs, kvs...)
	return e
}

// Trace will log formatted args as 'msg' field to the log at TRACE level.
//
//go:noinline
func (e Entry) Trace(a ...interface{}) {
	logf(e.ctx, 3, TRACE, e.kvs, args(len(a)), a...)
}

// Tracef will log format string as 'msg' field to the log at TRACE level.
//
//go:noinline
func (e Entry) Tracef(s string, a ...interface{}) {
	logf(e.ctx, 3, TRACE, e.kvs, s, a...)
}

// Debug will log formatted args as 'msg' field to the log at DEBUG level.
//
//go:noinline
func (e Entry) Debug(a ...interface{}) {
	logf(e.ctx, 3, DEBUG, e.kvs, args(len(a)), a...)
}

// Debugf will log format string as 'msg' field to the log at DEBUG level.
//
//go:noinline
func (e Entry) Debugf(s string, a ...interface{}) {
	logf(e.ctx, 3, DEBUG, e.kvs, s, a...)
}

// Info will log formatted args as 'msg' field to the log at INFO level.
//
//go:noinline
func (e Entry) Info(a ...interface{}) {
	logf(e.ctx, 3, INFO, e.kvs, args(len(a)), a...)
}

// Infof will log format string as 'msg' field to the log at INFO level.
//
//go:noinline
func (e Entry) Infof(s string, a ...interface{}) {
	logf(e.ctx, 3, INFO, e.kvs, s, a...)
}

// Warn will log formatted args as 'msg' field to the log at WARN level.
//
//go:noinline
func (e Entry) Warn(a ...interface{}) {
	logf(e.ctx, 3, WARN, e.kvs, args(len(a)), a...)
}

// Warnf will log format string as 'msg' field to the log at WARN level.
//
//go:noinline
func (e Entry) Warnf(s string, a ...interface{}) {
	logf(e.ctx, 3, WARN, e.kvs, s, a...)
}

// Error will log formatted args as 'msg' field to the log at ERROR level.
//
//go:noinline
func (e Entry) Error(a ...interface{}) {
	logf(e.ctx, 3, ERROR, e.kvs, args(len(a)), a...)
}

// Errorf will log format string as 'msg' field to the log at ERROR level.
//
//go:noinline
func (e Entry) Errorf(s string, a ...interface{}) {
	logf(e.ctx, 3, ERROR, e.kvs, s, a...)
}

// Panic will log formatted args as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func (e Entry) Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(e.ctx, 3, PANIC, e.kvs, args(len(a)), a...)
}

// Panicf will log format string as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func (e Entry) Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(e.ctx, 3, PANIC, e.kvs, s, a...)
}

// Log will log formatted args as 'msg' field to the log at given level.
//
//go:noinline
func (e Entry) Log(lvl LEVEL, a ...interface{}) {
	logf(e.ctx, 3, lvl, e.kvs, args(len(a)), a...)
}

// Logf will log format string as 'msg' field to the log at given level.
//
//go:noinline
func (e Entry) Logf(lvl LEVEL, s string, a ...interface{}) {
	logf(e.ctx, 3, lvl, e.kvs, s, a...)
}

// Print will log formatted args to the stdout log output.
//
//go:noinline
func (e Entry) Print(a ...interface{}) {
	printf(3, e.kvs, args(len(a)), a...)
}

// Printf will log format string to the stdout log output.
//
//go:noinline
func (e Entry) Printf(s string, a ...interface{}) {
	printf(3, e.kvs, s, a...)
}
