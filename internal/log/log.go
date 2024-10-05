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
	"slices"
	"strings"
	"time"

	"codeberg.org/gruf/go-kv"
)

var (
	// loglvl is the currently
	// set logging output
	loglvl LEVEL

	// lvlstrs is the lookup table
	// of all log levels to strings.
	lvlstrs = [int(ALL) + 1]string{
		TRACE: "TRACE",
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		PANIC: "PANIC",
	}

	// syslog output, only set if enabled.
	sysout *syslog.Writer

	// timefmt is the logging time format used, which includes
	// the full field and required quoting
	timefmt = `timestamp="02/01/2006 15:04:05.000" `

	// ctxhooks allows modifying log content based on context.
	ctxhooks []func(context.Context, []kv.Field) []kv.Field
)

// Hook adds the given hook to the global logger context hooks stack.
func Hook(hook func(ctx context.Context, kvs []kv.Field) []kv.Field) {
	ctxhooks = append(ctxhooks, hook)
}

// Level returns the currently set log
func Level() LEVEL {
	return loglvl
}

// SetLevel sets the max logging
func SetLevel(lvl LEVEL) {
	loglvl = lvl
}

// TimeFormat returns the currently-set timestamp format.
func TimeFormat() string {
	return timefmt
}

// SetTimeFormat sets the timestamp format to the given string.
func SetTimeFormat(format string) {
	if format == "" {
		timefmt = format
		return
	}
	timefmt = `timestamp="` + format + `" `
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

// Note that most of the below logging
// functions we specifically do NOT allow
// the Go buildchain to inline, to ensure
// expected behaviour in caller fetching.

// Trace will log formatted args as 'msg' field to the log at TRACE level.
//
//go:noinline
func Trace(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, TRACE, nil, args(len(a)), a...)
}

// Tracef will log format string as 'msg' field to the log at TRACE level.
//
//go:noinline
func Tracef(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, TRACE, nil, s, a...)
}

// TraceKV will log the one key-value field to the log at TRACE level.
//
//go:noinline
func TraceKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, TRACE, []kv.Field{{K: key, V: value}}, "")
}

// TraceKVs will log key-value fields to the log at TRACE level.
//
//go:noinline
func TraceKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, TRACE, kvs, "")
}

// Debug will log formatted args as 'msg' field to the log at DEBUG level.
//
//go:noinline
func Debug(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, DEBUG, nil, args(len(a)), a...)
}

// Debugf will log format string as 'msg' field to the log at DEBUG level.
//
//go:noinline
func Debugf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, DEBUG, nil, s, a...)
}

// DebugKV will log the one key-value field to the log at DEBUG level.
//
//go:noinline
func DebugKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, DEBUG, []kv.Field{{K: key, V: value}}, "")
}

// DebugKVs will log key-value fields to the log at DEBUG level.
//
//go:noinline
func DebugKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, DEBUG, kvs, "")
}

// Info will log formatted args as 'msg' field to the log at INFO level.
//
//go:noinline
func Info(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, INFO, nil, args(len(a)), a...)
}

// Infof will log format string as 'msg' field to the log at INFO level.
//
//go:noinline
func Infof(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, INFO, nil, s, a...)
}

// InfoKV will log the one key-value field to the log at INFO level.
//
//go:noinline
func InfoKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, INFO, []kv.Field{{K: key, V: value}}, "")
}

// InfoKVs will log key-value fields to the log at INFO level.
//
//go:noinline
func InfoKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, INFO, kvs, "")
}

// Warn will log formatted args as 'msg' field to the log at WARN level.
//
//go:noinline
func Warn(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, WARN, nil, args(len(a)), a...)
}

// Warnf will log format string as 'msg' field to the log at WARN level.
//
//go:noinline
func Warnf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, WARN, nil, s, a...)
}

// WarnKV will log the one key-value field to the log at WARN level.
//
//go:noinline
func WarnKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, WARN, []kv.Field{{K: key, V: value}}, "")
}

// WarnKVs will log key-value fields to the log at WARN level.
//
//go:noinline
func WarnKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, WARN, kvs, "")
}

// Error will log formatted args as 'msg' field to the log at ERROR level.
//
//go:noinline
func Error(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, ERROR, nil, args(len(a)), a...)
}

// Errorf will log format string as 'msg' field to the log at ERROR level.
//
//go:noinline
func Errorf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, ERROR, nil, s, a...)
}

// ErrorKV will log the one key-value field to the log at ERROR level.
//
//go:noinline
func ErrorKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, ERROR, []kv.Field{{K: key, V: value}}, "")
}

// ErrorKVs will log key-value fields to the log at ERROR level.
//
//go:noinline
func ErrorKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, ERROR, kvs, "")
}

// Panic will log formatted args as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func Panic(ctx context.Context, a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(ctx, 3, PANIC, nil, args(len(a)), a...)
}

// Panicf will log format string as 'msg' field to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func Panicf(ctx context.Context, s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(ctx, 3, PANIC, nil, s, a...)
}

// PanicKV will log the one key-value field to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func PanicKV(ctx context.Context, key string, value interface{}) {
	defer panic(kv.Field{K: key, V: value}.String())
	logf(ctx, 3, PANIC, []kv.Field{{K: key, V: value}}, "")
}

// PanicKVs will log key-value fields to the log at PANIC level.
// This will then call panic causing the application to crash.
//
//go:noinline
func PanicKVs(ctx context.Context, kvs ...kv.Field) {
	defer panic(kv.Fields(kvs).String())
	logf(ctx, 3, PANIC, kvs, "")
}

// Log will log formatted args as 'msg' field to the log at given level.
//
//go:noinline
func Log(ctx context.Context, lvl LEVEL, a ...interface{}) {
	logf(ctx, 3, lvl, nil, args(len(a)), a...)
}

// Logf will log format string as 'msg' field to the log at given level.
//
//go:noinline
func Logf(ctx context.Context, lvl LEVEL, s string, a ...interface{}) {
	logf(ctx, 3, lvl, nil, s, a...)
}

// LogKV will log the one key-value field to the log at given level.
//
//go:noinline
func LogKV(ctx context.Context, lvl LEVEL, key string, value interface{}) { //nolint:revive
	logf(ctx, 3, lvl, []kv.Field{{K: key, V: value}}, "")
}

// LogKVs will log key-value fields to the log at given level.
//
//go:noinline
func LogKVs(ctx context.Context, lvl LEVEL, kvs ...kv.Field) { //nolint:revive
	logf(ctx, 3, lvl, kvs, "")
}

// Print will log formatted args to the stdout log output.
//
//go:noinline
func Print(a ...interface{}) {
	printf(3, nil, args(len(a)), a...)
}

// Printf will log format string to the stdout log output.
//
//go:noinline
func Printf(s string, a ...interface{}) {
	printf(3, nil, s, a...)
}

// PrintKVs will log the one key-value field to the stdout log output.
//
//go:noinline
func PrintKV(key string, value interface{}) {
	printf(3, []kv.Field{{K: key, V: value}}, "")
}

// PrintKVs will log key-value fields to the stdout log output.
//
//go:noinline
func PrintKVs(kvs ...kv.Field) {
	printf(3, kvs, "")
}

//go:noinline
func printf(depth int, fields []kv.Field, s string, a ...interface{}) {
	// Acquire buffer
	buf := getBuf()

	// Append formatted timestamp according to `timefmt`
	buf.B = time.Now().AppendFormat(buf.B, timefmt)

	// Append formatted caller func
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, Caller(depth+1)...)
	buf.B = append(buf.B, ' ')

	if len(fields) > 0 {
		// Append formatted fields
		kv.Fields(fields).AppendFormat(buf, false)
		buf.B = append(buf.B, ' ')
	}

	// Append formatted args
	fmt.Fprintf(buf, s, a...)

	if buf.B[len(buf.B)-1] != '\n' {
		// Append a final newline
		buf.B = append(buf.B, '\n')
	}

	if sysout != nil {
		// Write log entry to syslog
		logsys(INFO, buf.String())
	}

	// Write to log and release
	_, _ = os.Stdout.Write(buf.B)
	putBuf(buf)
}

//go:noinline
func logf(ctx context.Context, depth int, lvl LEVEL, fields []kv.Field, s string, a ...interface{}) {
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

	// Acquire buffer
	buf := getBuf()

	// Append formatted timestamp according to `timefmt`
	buf.B = time.Now().AppendFormat(buf.B, timefmt)

	// Append formatted caller func
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, Caller(depth+1)...)
	buf.B = append(buf.B, ' ')

	// Append formatted level string
	buf.B = append(buf.B, `level=`...)
	buf.B = append(buf.B, lvlstrs[lvl]...)
	buf.B = append(buf.B, ' ')

	if ctx != nil {
		// Pass context through hooks.
		for _, hook := range ctxhooks {
			fields = hook(ctx, fields)
		}
	}

	if s != "" {
		// Append message to log fields.
		fields = slices.Grow(fields, 1)
		fields = append(fields, kv.Field{
			K: "msg", V: fmt.Sprintf(s, a...),
		})
	}

	// Append formatted fields to log buffer.
	kv.Fields(fields).AppendFormat(buf, false)

	if buf.B[len(buf.B)-1] != '\n' {
		// Append a final newline
		buf.B = append(buf.B, '\n')
	}

	if sysout != nil {
		// Write log entry to syslog
		logsys(lvl, buf.String())
	}

	// Write to log and release
	_, _ = out.Write(buf.B)
	putBuf(buf)
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

// args returns an args format string of format '%v' * count.
func args(count int) string {
	const args = `%v%v%v%v%v%v%v%v%v%v` +
		`%v%v%v%v%v%v%v%v%v%v` +
		`%v%v%v%v%v%v%v%v%v%v` +
		`%v%v%v%v%v%v%v%v%v%v`

	// Use predetermined args str
	if count < len(args) {
		return args[:count*2]
	}

	// Allocate buffer of needed len
	var buf strings.Builder
	buf.Grow(count * 2)

	// Manually build an args str
	for i := 0; i < count; i++ {
		buf.WriteString(`%v`)
	}

	return buf.String()
}
