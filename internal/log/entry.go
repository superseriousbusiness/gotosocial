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

package log

import (
	"context"
	"fmt"
	"syscall"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

type Entry struct {
	ctx context.Context
	kvs []kv.Field
}

func (e Entry) WithContext(ctx context.Context) Entry {
	e.ctx = ctx
	return e
}

func (e Entry) WithField(key string, value interface{}) Entry {
	e.kvs = append(e.kvs, kv.Field{K: key, V: value})
	return e
}

func (e Entry) WithFields(kvs ...kv.Field) Entry {
	e.kvs = append(e.kvs, kvs...)
	return e
}

func (e Entry) Trace(a ...interface{}) {
	logf(e.ctx, 3, level.TRACE, e.kvs, args(len(a)), a...)
}

func (e Entry) Tracef(s string, a ...interface{}) {
	logf(e.ctx, 3, level.TRACE, e.kvs, s, a...)
}

func (e Entry) Debug(a ...interface{}) {
	logf(e.ctx, 3, level.DEBUG, e.kvs, args(len(a)), a...)
}

func (e Entry) Debugf(s string, a ...interface{}) {
	logf(e.ctx, 3, level.DEBUG, e.kvs, s, a...)
}

func (e Entry) Info(a ...interface{}) {
	logf(e.ctx, 3, level.INFO, e.kvs, args(len(a)), a...)
}

func (e Entry) Infof(s string, a ...interface{}) {
	logf(e.ctx, 3, level.INFO, e.kvs, s, a...)
}

func (e Entry) Warn(a ...interface{}) {
	logf(e.ctx, 3, level.WARN, e.kvs, args(len(a)), a...)
}

func (e Entry) Warnf(s string, a ...interface{}) {
	logf(e.ctx, 3, level.WARN, e.kvs, s, a...)
}

func (e Entry) Error(a ...interface{}) {
	logf(e.ctx, 3, level.ERROR, e.kvs, args(len(a)), a...)
}

func (e Entry) Errorf(s string, a ...interface{}) {
	logf(e.ctx, 3, level.ERROR, e.kvs, s, a...)
}

func (e Entry) Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	logf(e.ctx, 3, level.FATAL, e.kvs, args(len(a)), a...)
}

func (e Entry) Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(e.ctx, 3, level.FATAL, e.kvs, s, a...)
}

func (e Entry) Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(e.ctx, 3, level.PANIC, e.kvs, args(len(a)), a...)
}

func (e Entry) Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(e.ctx, 3, level.PANIC, e.kvs, s, a...)
}

func (e Entry) Log(lvl level.LEVEL, a ...interface{}) {
	logf(e.ctx, 3, lvl, e.kvs, args(len(a)), a...)
}

func (e Entry) Logf(lvl level.LEVEL, s string, a ...interface{}) {
	logf(e.ctx, 3, lvl, e.kvs, s, a...)
}

func (e Entry) Print(a ...interface{}) {
	printf(3, e.kvs, args(len(a)), a...)
}

func (e Entry) Printf(s string, a ...interface{}) {
	printf(3, e.kvs, s, a...)
}
