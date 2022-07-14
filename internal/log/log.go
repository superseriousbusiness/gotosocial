package log

import (
	"log/syslog"
	"os"
	"syscall"

	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2"
	"codeberg.org/gruf/go-logger/v2/entry"
	"codeberg.org/gruf/go-logger/v2/level"
)

var (

	// Default log config with field formatter.
	cfg = logger.Config{
		Format: entry.NewFieldFormatter(nil),
	}

	// Default logger flags (includes timestamp).
	flags = logger.Flags(0).SetTime()

	// Preprepared stdout/stderr logs, customized in Initialize().
	stdout = logger.NewWith(os.Stdout, cfg, 0, flags)
	stderr = logger.NewWith(os.Stderr, cfg, 0, flags)

	// Syslog output, only set if enabled.
	sysout *syslog.Writer
)

func With(key string, value interface{}) Entry {
	return Entry{fields: []kv.Field{{K: key, V: value}}}
}

func WithFields(fields ...kv.Field) Entry {
	return Entry{fields: fields}
}

func Trace(a ...interface{}) {
	log(level.TRACE, nil, a...)
}

func Tracef(s string, a ...interface{}) {
	logf(level.TRACE, nil, s, a...)
}

func Debug(a ...interface{}) {
	log(level.DEBUG, nil, a...)
}

func Debugf(s string, a ...interface{}) {
	logf(level.DEBUG, nil, s, a...)
}

func Info(a ...interface{}) {
	log(level.INFO, nil, a...)
}

func Infof(s string, a ...interface{}) {
	logf(level.INFO, nil, s, a...)
}

func Warn(a ...interface{}) {
	log(level.WARN, nil, a...)
}

func Warnf(s string, a ...interface{}) {
	logf(level.WARN, nil, s, a...)
}

func Error(a ...interface{}) {
	log(level.ERROR, nil, a...)
}

func Errorf(s string, a ...interface{}) {
	logf(level.ERROR, nil, s, a...)
}

func Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	log(level.FATAL, nil, a...)
}

func Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, nil, s, a...)
}

func log(lvl level.LEVEL, fields []kv.Field, a ...interface{}) {
	var out *logger.Logger

	if lvl <= level.ERROR {
		out = stderr
	} else {
		out = stdout
	}

	// Acquire entry from pool
	entry := out.Entry(4)

	// Write formatted entry
	entry.Timestamp()
	entry.WithLevel(lvl)
	entry.Fields(fields...)
	entry.Msg(a...)

	if sysout != nil {
		// Log this entry to syslog
		logsys(lvl, entry.String())
	}

	// Write to main log
	out.Write(entry)
}

func logf(lvl level.LEVEL, fields []kv.Field, s string, a ...interface{}) {
	var out *logger.Logger

	if lvl <= level.ERROR {
		out = stderr
	} else {
		out = stdout
	}

	// Acquire entry from pool
	entry := out.Entry(4)

	// Write formatted entry
	entry.Timestamp()
	entry.WithLevel(lvl)
	entry.Fields(fields...)
	entry.Msgf(s, a...)

	if sysout != nil {
		// Log this entry to syslog
		logsys(lvl, entry.String())
	}

	// Write to main log
	out.Write(entry)
}

func logsys(lvl level.LEVEL, msg string) {
	switch lvl {
	case level.DEBUG:
		sysout.Debug(msg)
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
