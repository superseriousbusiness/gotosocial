/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"io"
	"log/syslog"
	"os"
	"strings"
	"syscall"
	"time"

	"codeberg.org/gruf/go-atomics"
	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

var (
	// loglvl is the currently set logging level.
	loglvl atomics.Uint32

	// lvlstrs is the lookup table of log levels to strings.
	lvlstrs = level.Default()

	// Preprepared stdout/stderr log writers.
	stdout = &safewriter{w: os.Stdout}
	stderr = &safewriter{w: os.Stderr}

	// Syslog output, only set if enabled.
	sysout *syslog.Writer
)

// Level returns the currently set log level.
func Level() level.LEVEL {
	return level.LEVEL(loglvl.Load())
}

// SetLevel sets the max logging level.
func SetLevel(lvl level.LEVEL) {
	loglvl.Store(uint32(lvl))
}

func WithField(key string, value interface{}) Entry {
	return Entry{fields: []kv.Field{{K: key, V: value}}}
}

func WithFields(fields ...kv.Field) Entry {
	return Entry{fields: fields}
}

func Trace(a ...interface{}) {
	logf(level.TRACE, nil, args(len(a)), a...)
}

func Tracef(s string, a ...interface{}) {
	logf(level.TRACE, nil, s, a...)
}

func Debug(a ...interface{}) {
	logf(level.DEBUG, nil, args(len(a)), a...)
}

func Debugf(s string, a ...interface{}) {
	logf(level.DEBUG, nil, s, a...)
}

func Info(a ...interface{}) {
	logf(level.INFO, nil, args(len(a)), a...)
}

func Infof(s string, a ...interface{}) {
	logf(level.INFO, nil, s, a...)
}

func Warn(a ...interface{}) {
	logf(level.WARN, nil, args(len(a)), a...)
}

func Warnf(s string, a ...interface{}) {
	logf(level.WARN, nil, s, a...)
}

func Error(a ...interface{}) {
	logf(level.ERROR, nil, args(len(a)), a...)
}

func Errorf(s string, a ...interface{}) {
	logf(level.ERROR, nil, s, a...)
}

func Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, nil, args(len(a)), a...)
}

func Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, nil, s, a...)
}

func Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(level.PANIC, nil, args(len(a)), a...)
}

func Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(level.PANIC, nil, s, a...)
}

// Log will log formatted args as 'msg' field to the log at given level.
func Log(lvl level.LEVEL, a ...interface{}) {
	logf(lvl, nil, args(len(a)), a...)
}

// Logf will log format string as 'msg' field to the log at given level.
func Logf(lvl level.LEVEL, s string, a ...interface{}) {
	logf(lvl, nil, s, a...)
}

// Print will log formatted args to the stdout log output.
func Print(a ...interface{}) {
	printf(nil, args(len(a)), a...)
}

// Print will log format string to the stdout log output.
func Printf(s string, a ...interface{}) {
	printf(nil, s, a...)
}

func printf(fields []kv.Field, s string, a ...interface{}) {
	// Acquire buffer
	buf := getBuf()

	// Append formatted timestamp
	now := time.Now().Format("02/01/2006 15:04:05.000")
	buf.B = append(buf.B, `timestamp="`...)
	buf.B = append(buf.B, now...)
	buf.B = append(buf.B, `" `...)

	// Append formatted caller func
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, caller(3)...)
	buf.B = append(buf.B, ' ')

	if len(fields) > 0 {
		// Append formatted fields
		kv.Fields(fields).AppendFormat(buf)
		buf.B = append(buf.B, ' ')
	}

	// Append formatted args
	fmt.Fprintf(buf, s, a...)

	// Append a final newline
	buf.B = append(buf.B, '\n')

	// Write to log and release
	_, _ = stdout.Write(buf.B)
	putBuf(buf)
}

func logf(lvl level.LEVEL, fields []kv.Field, s string, a ...interface{}) {
	var out io.Writer

	// Check if enabled.
	if lvl > Level() {
		return
	}

	// Split errors to stderr,
	// all else goes to stdout.
	if lvl <= level.ERROR {
		out = stderr
	} else {
		out = stdout
	}

	// Acquire buffer
	buf := getBuf()

	// Append formatted timestamp
	now := time.Now().Format("02/01/2006 15:04:05.000")
	buf.B = append(buf.B, `timestamp="`...)
	buf.B = append(buf.B, now...)
	buf.B = append(buf.B, `" `...)

	// Append formatted caller func
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, caller(3)...)
	buf.B = append(buf.B, ' ')

	// Append formatted level string
	buf.B = append(buf.B, `level=`...)
	buf.B = append(buf.B, lvlstrs[lvl]...)
	buf.B = append(buf.B, ' ')

	// Append formatted fields with msg
	kv.Fields(append(fields, kv.Field{
		K: "msg", V: fmt.Sprintf(s, a...),
	})).AppendFormat(buf)

	// Append a final newline
	buf.B = append(buf.B, '\n')

	if sysout != nil {
		// Write log entry to syslog
		logsys(lvl, buf.String())
	}

	// Write to log and release
	_, _ = out.Write(buf.B)
	putBuf(buf)
}

// logsys will log given msg at given severity to the syslog.
func logsys(lvl level.LEVEL, msg string) {
	// Truncate message if > 1700 chars
	if len(msg) > 1700 {
		msg = msg[:1697] + "..."
	}

	// Log at appropriate syslog severity
	switch lvl {
	case level.TRACE:
	case level.DEBUG:
	case level.INFO:
		sysout.Info(msg)
	case level.WARN:
		sysout.Warning(msg)
	case level.ERROR:
		sysout.Err(msg)
	case level.FATAL:
		sysout.Crit(msg)
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
