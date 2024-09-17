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
	"syscall"
	"time"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

var (
	// loglvl is the currently set logging level.
	loglvl level.LEVEL

	// lvlstrs is the lookup table of log levels to strings.
	lvlstrs = level.Default()

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

// Level returns the currently set log level.
func Level() level.LEVEL {
	return loglvl
}

// SetLevel sets the max logging level.
func SetLevel(lvl level.LEVEL) {
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

func WithContext(ctx context.Context) Entry {
	return Entry{ctx: ctx}
}

func WithField(key string, value interface{}) Entry {
	return New().WithField(key, value)
}

func WithFields(fields ...kv.Field) Entry {
	return New().WithFields(fields...)
}

func Trace(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, level.TRACE, nil, args(len(a)), a...)
}

func Tracef(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, level.TRACE, nil, s, a...)
}

func TraceKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.TRACE, []kv.Field{{K: key, V: value}}, "")
}

func TraceKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.TRACE, kvs, "")
}

func Debug(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, level.DEBUG, nil, args(len(a)), a...)
}

func Debugf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, level.DEBUG, nil, s, a...)
}

func DebugKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.DEBUG, []kv.Field{{K: key, V: value}}, "")
}

func DebugKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.DEBUG, kvs, "")
}

func Info(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, level.INFO, nil, args(len(a)), a...)
}

func Infof(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, level.INFO, nil, s, a...)
}

func InfoKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.INFO, []kv.Field{{K: key, V: value}}, "")
}

func InfoKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.INFO, kvs, "")
}

func Warn(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, level.WARN, nil, args(len(a)), a...)
}

func Warnf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, level.WARN, nil, s, a...)
}

func WarnKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.WARN, []kv.Field{{K: key, V: value}}, "")
}

func WarnKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.WARN, kvs, "")
}

func Error(ctx context.Context, a ...interface{}) {
	logf(ctx, 3, level.ERROR, nil, args(len(a)), a...)
}

func Errorf(ctx context.Context, s string, a ...interface{}) {
	logf(ctx, 3, level.ERROR, nil, s, a...)
}

func ErrorKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.ERROR, []kv.Field{{K: key, V: value}}, "")
}

func ErrorKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.WARN, kvs, "")
}

func Fatal(ctx context.Context, a ...interface{}) {
	defer syscall.Exit(1)
	logf(ctx, 3, level.FATAL, nil, args(len(a)), a...)
}

func Fatalf(ctx context.Context, s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(ctx, 3, level.FATAL, nil, s, a...)
}

func FatalKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.FATAL, []kv.Field{{K: key, V: value}}, "")
}

func FatalKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.FATAL, kvs, "")
}

func Panic(ctx context.Context, a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(ctx, 3, level.PANIC, nil, args(len(a)), a...)
}

func Panicf(ctx context.Context, s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(ctx, 3, level.PANIC, nil, s, a...)
}

func PanicKV(ctx context.Context, key string, value interface{}) {
	logf(ctx, 3, level.PANIC, []kv.Field{{K: key, V: value}}, "")
}

func PanicKVs(ctx context.Context, kvs ...kv.Field) {
	logf(ctx, 3, level.PANIC, kvs, "")
}

// Log will log formatted args as 'msg' field to the log at given level.
func Log(ctx context.Context, lvl level.LEVEL, a ...interface{}) {
	logf(ctx, 3, lvl, nil, args(len(a)), a...)
}

// Logf will log format string as 'msg' field to the log at given level.
func Logf(ctx context.Context, lvl level.LEVEL, s string, a ...interface{}) {
	logf(ctx, 3, lvl, nil, s, a...)
}

// LogKV will log the one key-value field to the log at given level.
func LogKV(ctx context.Context, lvl level.LEVEL, key string, value interface{}) { //nolint:revive
	logf(ctx, 3, level.DEBUG, []kv.Field{{K: key, V: value}}, "")
}

// LogKVs will log key-value fields to the log at given level.
func LogKVs(ctx context.Context, lvl level.LEVEL, kvs ...kv.Field) { //nolint:revive
	logf(ctx, 3, lvl, kvs, "")
}

// Print will log formatted args to the stdout log output.
func Print(a ...interface{}) {
	printf(3, nil, args(len(a)), a...)
}

// Printf will log format string to the stdout log output.
func Printf(s string, a ...interface{}) {
	printf(3, nil, s, a...)
}

// PrintKVs will log the one key-value field to the stdout log output.
func PrintKV(key string, value interface{}) {
	printf(3, []kv.Field{{K: key, V: value}}, "")
}

// PrintKVs will log key-value fields to the stdout log output.
func PrintKVs(kvs ...kv.Field) {
	printf(3, kvs, "")
}

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
		logsys(level.INFO, buf.String())
	}

	// Write to log and release
	_, _ = os.Stdout.Write(buf.B)
	putBuf(buf)
}

func logf(ctx context.Context, depth int, lvl level.LEVEL, fields []kv.Field, s string, a ...interface{}) {
	var out *os.File

	// Check if enabled.
	if lvl > Level() {
		return
	}

	// Split errors to stderr,
	// all else goes to stdout.
	if lvl <= level.ERROR {
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
func logsys(lvl level.LEVEL, msg string) {
	if max := 2048; len(msg) > max {
		// Truncate up to max
		msg = msg[:max]
	}
	switch lvl {
	case level.TRACE, level.DEBUG:
		_ = sysout.Debug(msg)
	case level.INFO:
		_ = sysout.Info(msg)
	case level.WARN:
		_ = sysout.Warning(msg)
	case level.ERROR:
		_ = sysout.Err(msg)
	case level.FATAL, level.PANIC:
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
