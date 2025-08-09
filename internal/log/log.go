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
	"log/syslog"
	"os"
	"runtime"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/log/format"
	"code.superseriousbusiness.org/gotosocial/internal/log/level"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"codeberg.org/gruf/go-kv/v2"
)

var (
	// loglvl is the currently
	// set logging output.
	loglvl = level.UNSET

	// appendFormat stores log
	// entry formatting function.
	appendFormat = (&format.Logfmt{
		Base: format.Base{TimeFormat: timefmt},
	}).Format

	// syslog output, only set if enabled.
	sysout *syslog.Writer

	// timefmt is the logging time format used.
	timefmt = `02/01/2006 15:04:05.000`

	// ctxhooks allows modifying log content based on context.
	ctxhooks []func(context.Context, []kv.Field) []kv.Field
)

// Hook adds the given hook to the global logger context hooks stack.
func Hook(hook func(ctx context.Context, kvs []kv.Field) []kv.Field) {
	ctxhooks = append(ctxhooks, hook)
}

// Level returns the currently set log.
func Level() LEVEL {
	return loglvl
}

// SetLevel sets the max logging.
func SetLevel(lvl LEVEL) {
	loglvl = lvl
}

// TimeFormat returns the currently-set timestamp format.
func TimeFormat() string {
	return timefmt
}

// SetTimeFormat sets the timestamp format to the given string.
func SetTimeFormat(format string) {
	timefmt = format
}

// SetJSON enables / disables JSON log output formatting.
func SetJSON(enabled bool) {
	if enabled {
		var fmt format.JSON
		fmt.TimeFormat = timefmt
		appendFormat = fmt.Format
	} else {
		var fmt format.Logfmt
		fmt.TimeFormat = timefmt
		appendFormat = fmt.Format
	}
}

// New starts a new log entry.
func New() Entry {
	return Entry{}
}

// WithContext returns a new prepared Entry{} with context.
func WithContext(ctx context.Context) Entry {
	return Entry{ctx: ctx}
}

// WithField returns a new prepared Entry{} with key-value field.
func WithField(key string, value interface{}) Entry {
	return Entry{kvs: []kv.Field{{K: key, V: value}}}
}

// WithFields returns a new prepared Entry{} with key-value fields.
func WithFields(fields ...kv.Field) Entry {
	return Entry{kvs: fields}
}

// Trace will log formatted args as 'msg' field to the log at TRACE level.
func Trace(ctx context.Context, a ...interface{}) {
	logf(ctx, TRACE, nil, "", a...)
}

// Tracef will log format string as 'msg' field to the log at TRACE level.
func Tracef(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, TRACE, nil, s, a...)
}

// TraceKV will log the one key-value field to the log at TRACE level.
func TraceKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, TRACE, []kv.Field{{K: key, V: value}}, "")
}

// TraceKVs will log key-value fields to the log at TRACE level.
func TraceKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, TRACE, kvs, "")
}

// Debug will log formatted args as 'msg' field to the log at DEBUG level.
func Debug(ctx context.Context, a ...interface{}) {
	logf(ctx, DEBUG, nil, "", a...)
}

// Debugf will log format string as 'msg' field to the log at DEBUG level.
func Debugf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, DEBUG, nil, s, a...)
}

// DebugKV will log the one key-value field to the log at DEBUG level.
func DebugKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, DEBUG, []kv.Field{{K: key, V: value}}, "")
}

// DebugKVs will log key-value fields to the log at DEBUG level.
func DebugKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, DEBUG, kvs, "")
}

// Info will log formatted args as 'msg' field to the log at INFO level.
func Info(ctx context.Context, a ...interface{}) {
	logf(ctx, INFO, nil, "", a...)
}

// Infof will log format string as 'msg' field to the log at INFO level.
func Infof(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, INFO, nil, s, a...)
}

// InfoKV will log the one key-value field to the log at INFO level.
func InfoKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, INFO, []kv.Field{{K: key, V: value}}, "")
}

// InfoKVs will log key-value fields to the log at INFO level.
func InfoKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, INFO, kvs, "")
}

// Warn will log formatted args as 'msg' field to the log at WARN level.
func Warn(ctx context.Context, a ...interface{}) {
	logf(ctx, WARN, nil, "", a...)
}

// Warnf will log format string as 'msg' field to the log at WARN level.
func Warnf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, WARN, nil, s, a...)
}

// WarnKV will log the one key-value field to the log at WARN level.
func WarnKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, WARN, []kv.Field{{K: key, V: value}}, "")
}

// WarnKVs will log key-value fields to the log at WARN level.
func WarnKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, WARN, kvs, "")
}

// Error will log formatted args as 'msg' field to the log at ERROR level.
func Error(ctx context.Context, a ...interface{}) {
	logf(ctx, ERROR, nil, "", a...)
}

// Errorf will log format string as 'msg' field to the log at ERROR level.
func Errorf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, ERROR, nil, s, a...)
}

// ErrorKV will log the one key-value field to the log at ERROR level.
func ErrorKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, ERROR, []kv.Field{{K: key, V: value}}, "")
}

// ErrorKVs will log key-value fields to the log at ERROR level.
func ErrorKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, ERROR, kvs, "")
}

// Panic will log formatted args as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
func Panic(ctx context.Context, a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(ctx, PANIC, nil, "", a...)
}

// Panicf will log format string as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
func Panicf(ctx context.Context, s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(ctx, PANIC, nil, s, a...)
}

// PanicKV will log the one key-value field to the log at PANIC level.
// This will then call panic causing the application to crash.
func PanicKV(ctx context.Context, key string, value interface{}) {
	defer panic(kv.Field{K: key, V: value}.String())
	logf(ctx, PANIC, []kv.Field{{K: key, V: value}}, "")
}

// PanicKVs will log key-value fields to the log at PANIC level.
// This will then call panic causing the application to crash.
func PanicKVs(ctx context.Context, kvs ...kv.Field) {
	defer panic(kv.Fields(kvs).String())
	logf(ctx, PANIC, kvs, "")
}

// Log will log formatted args as 'msg' field to the log at given level.
func Log(ctx context.Context, lvl LEVEL, a ...interface{}) { //nolint:revive
	logf(ctx, lvl, nil, "", a...)
}

// Logf will log format string as 'msg' field to the log at given level.
func Logf(ctx context.Context, lvl LEVEL, s string, a ...interface{}) { //nolint:revive
	logf(ctx, lvl, nil, s, a...)
}

// LogKV will log the one key-value field to the log at given level.
func LogKV(ctx context.Context, lvl LEVEL, key string, value interface{}) { //nolint:revive
	logf(ctx, lvl, []kv.Field{{K: key, V: value}}, "")
}

// LogKVs will log key-value fields to the log at given level.
func LogKVs(ctx context.Context, lvl LEVEL, kvs ...kv.Field) { //nolint:revive
	logf(ctx, lvl, kvs, "")
}

// Print will log formatted args to the stdout log output.
func Print(a ...interface{}) {
	logf(context.Background(), UNSET, nil, "", a...)
}

// Printf will log format string to the stdout log output.
func Printf(s string, a ...interface{}) {
	logf(context.Background(), UNSET, nil, s, a...)
}

//go:noinline
func logf(ctx context.Context, lvl LEVEL, fields []kv.Field, msg string, args ...interface{}) {
	var out *os.File

	// Check if enabled.
	if lvl > loglvl {
		return
	}

	// Split errors to stderr,
	// all else goes to stdout.
	if lvl <= ERROR {
		out = os.Stderr
	} else {
		out = os.Stdout
	}

	// Get log stamp.
	now := time.Now()

	// Get caller information.
	pcs := make([]uintptr, 1)
	_ = runtime.Callers(3, pcs)

	// Acquire buffer.
	buf := getBuf()
	defer putBuf(buf)

	if ctx != nil {
		// Ensure fields have space for context hooks.
		fields = xslices.GrowJust(fields, len(ctxhooks))

		// Pass context through hooks.
		for _, hook := range ctxhooks {
			fields = hook(ctx, fields)
		}
	}

	// If no args, use placeholders.
	if msg == "" && len(args) > 0 {
		const argstr = `%v%v%v%v%v%v%v%v%v%v` +
			`%v%v%v%v%v%v%v%v%v%v` +
			`%v%v%v%v%v%v%v%v%v%v` +
			`%v%v%v%v%v%v%v%v%v%v`
		msg = argstr[:2*len(args)]
	}

	if msg != "" {
		// Format the message string.
		msg = fmt.Sprintf(msg, args...)
	}

	// Append formatted
	// entry to buffer.
	appendFormat(buf,
		now,
		pcs[0],
		lvl,
		fields,
		msg,
	)

	// Ensure a final new-line char.
	if buf.B[len(buf.B)-1] != '\n' {
		buf.B = append(buf.B, '\n')
	}

	if sysout != nil {
		// Write log entry to syslog
		logsys(lvl, buf.String())
	}

	// Write to output file.
	_, _ = out.Write(buf.B)
}

// logsys will log given msg at given severity to the syslog.
// Max length: https://www.rfc-editor.org/rfc/rfc5424.html#section-6.1
func logsys(lvl LEVEL, msg string) {
	if max := 2048; len(msg) > max {
		// Truncate up to max
		msg = msg[:max]
	}
	switch lvl {
	case TRACE, DEBUG:
		_ = sysout.Debug(msg)
	case INFO:
		_ = sysout.Info(msg)
	case WARN:
		_ = sysout.Warning(msg)
	case ERROR:
		_ = sysout.Err(msg)
	case PANIC:
		_ = sysout.Crit(msg)
	}
}
