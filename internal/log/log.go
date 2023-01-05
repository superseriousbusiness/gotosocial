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
	"log/syslog"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

var (
	// loglvl is the currently set logging level.
	loglvl atomic.Uint32

	// lvlstrs is the lookup table of log levels to strings.
	lvlstrs = level.Default()

	// Syslog output, only set if enabled.
	sysout *syslog.Writer

	// timefmt is the logging time format used.
	timefmt = "02/01/2006 15:04:05.000"
)

// Level returns the currently set log level.
func Level() level.LEVEL {
	return level.LEVEL(loglvl.Load())
}

// SetLevel sets the max logging level.
func SetLevel(lvl level.LEVEL) {
	loglvl.Store(uint32(lvl))
}

// New starts a new log entry.
func New() Entry {
	return Entry{}
}

func WithField(key string, value interface{}) Entry {
	return Entry{fields: []kv.Field{{K: key, V: value}}}
}

func WithFields(fields ...kv.Field) Entry {
	return Entry{fields: fields}
}

func Trace(a ...interface{}) {
	logf(3, level.TRACE, nil, args(len(a)), a...)
}

func Tracef(s string, a ...interface{}) {
	logf(3, level.TRACE, nil, s, a...)
}

func Debug(a ...interface{}) {
	logf(3, level.DEBUG, nil, args(len(a)), a...)
}

func Debugf(s string, a ...interface{}) {
	logf(3, level.DEBUG, nil, s, a...)
}

func Info(a ...interface{}) {
	logf(3, level.INFO, nil, args(len(a)), a...)
}

func Infof(s string, a ...interface{}) {
	logf(3, level.INFO, nil, s, a...)
}

func Warn(a ...interface{}) {
	logf(3, level.WARN, nil, args(len(a)), a...)
}

func Warnf(s string, a ...interface{}) {
	logf(3, level.WARN, nil, s, a...)
}

func Error(a ...interface{}) {
	logf(3, level.ERROR, nil, args(len(a)), a...)
}

func Errorf(s string, a ...interface{}) {
	logf(3, level.ERROR, nil, s, a...)
}

func Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	logf(3, level.FATAL, nil, args(len(a)), a...)
}

func Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(3, level.FATAL, nil, s, a...)
}

func Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(3, level.PANIC, nil, args(len(a)), a...)
}

func Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(3, level.PANIC, nil, s, a...)
}

// Log will log formatted args as 'msg' field to the log at given level.
func Log(lvl level.LEVEL, a ...interface{}) {
	logf(3, lvl, nil, args(len(a)), a...)
}

// Logf will log format string as 'msg' field to the log at given level.
func Logf(lvl level.LEVEL, s string, a ...interface{}) {
	logf(3, lvl, nil, s, a...)
}

// Print will log formatted args to the stdout log output.
func Print(a ...interface{}) {
	printf(3, nil, args(len(a)), a...)
}

// Print will log format string to the stdout log output.
func Printf(s string, a ...interface{}) {
	printf(3, nil, s, a...)
}

func printf(depth int, fields []kv.Field, s string, a ...interface{}) {
	// Acquire buffer
	buf := getBuf()

	// Append formatted timestamp
	buf.B = append(buf.B, `timestamp="`...)
	buf.B = time.Now().AppendFormat(buf.B, timefmt)
	buf.B = append(buf.B, `" `...)

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

func logf(depth int, lvl level.LEVEL, fields []kv.Field, s string, a ...interface{}) {
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

	// Append formatted timestamp
	buf.B = append(buf.B, `timestamp="`...)
	buf.B = time.Now().AppendFormat(buf.B, timefmt)
	buf.B = append(buf.B, `" `...)

	// Append formatted caller func
	buf.B = append(buf.B, `func=`...)
	buf.B = append(buf.B, Caller(depth+1)...)
	buf.B = append(buf.B, ' ')

	// Append formatted level string
	buf.B = append(buf.B, `level=`...)
	buf.B = append(buf.B, lvlstrs[lvl]...)
	buf.B = append(buf.B, ' ')

	// Append formatted fields with msg
	kv.Fields(append(fields, kv.Field{
		K: "msg", V: fmt.Sprintf(s, a...),
	})).AppendFormat(buf, false)

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
