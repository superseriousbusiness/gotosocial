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
	"fmt"
	"syscall"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

type Entry struct {
	fields []kv.Field
}

func (e Entry) WithField(key string, value interface{}) Entry {
	e.fields = append(e.fields, kv.Field{K: key, V: value})
	return e
}

func (e Entry) WithFields(fields ...kv.Field) Entry {
	e.fields = append(e.fields, fields...)
	return e
}

func (e Entry) Trace(a ...interface{}) {
	logf(3, level.TRACE, e.fields, args(len(a)), a...)
}

func (e Entry) Tracef(s string, a ...interface{}) {
	logf(3, level.TRACE, e.fields, s, a...)
}

func (e Entry) Debug(a ...interface{}) {
	logf(3, level.DEBUG, e.fields, args(len(a)), a...)
}

func (e Entry) Debugf(s string, a ...interface{}) {
	logf(3, level.DEBUG, e.fields, s, a...)
}

func (e Entry) Info(a ...interface{}) {
	logf(3, level.INFO, e.fields, args(len(a)), a...)
}

func (e Entry) Infof(s string, a ...interface{}) {
	logf(3, level.INFO, e.fields, s, a...)
}

func (e Entry) Warn(a ...interface{}) {
	logf(3, level.WARN, e.fields, args(len(a)), a...)
}

func (e Entry) Warnf(s string, a ...interface{}) {
	logf(3, level.WARN, e.fields, s, a...)
}

func (e Entry) Error(a ...interface{}) {
	logf(3, level.ERROR, e.fields, args(len(a)), a...)
}

func (e Entry) Errorf(s string, a ...interface{}) {
	logf(3, level.ERROR, e.fields, s, a...)
}

func (e Entry) Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	logf(3, level.FATAL, e.fields, args(len(a)), a...)
}

func (e Entry) Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(3, level.FATAL, e.fields, s, a...)
}

func (e Entry) Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(3, level.PANIC, e.fields, args(len(a)), a...)
}

func (e Entry) Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(3, level.PANIC, e.fields, s, a...)
}

func (e Entry) Log(lvl level.LEVEL, a ...interface{}) {
	logf(3, lvl, e.fields, args(len(a)), a...)
}

func (e Entry) Logf(lvl level.LEVEL, s string, a ...interface{}) {
	logf(3, lvl, e.fields, s, a...)
}

func (e Entry) Print(a ...interface{}) {
	printf(3, e.fields, args(len(a)), a...)
}

func (e Entry) Printf(s string, a ...interface{}) {
	printf(3, e.fields, s, a...)
}
