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
	logf(level.TRACE, e.fields, args(len(a)), a...)
}

func (e Entry) Tracef(s string, a ...interface{}) {
	logf(level.TRACE, e.fields, s, a...)
}

func (e Entry) Debug(a ...interface{}) {
	logf(level.DEBUG, e.fields, args(len(a)), a...)
}

func (e Entry) Debugf(s string, a ...interface{}) {
	logf(level.DEBUG, e.fields, s, a...)
}

func (e Entry) Info(a ...interface{}) {
	logf(level.INFO, e.fields, args(len(a)), a...)
}

func (e Entry) Infof(s string, a ...interface{}) {
	logf(level.INFO, e.fields, s, a...)
}

func (e Entry) Warn(a ...interface{}) {
	logf(level.WARN, e.fields, args(len(a)), a...)
}

func (e Entry) Warnf(s string, a ...interface{}) {
	logf(level.WARN, e.fields, s, a...)
}

func (e Entry) Error(a ...interface{}) {
	logf(level.ERROR, e.fields, args(len(a)), a...)
}

func (e Entry) Errorf(s string, a ...interface{}) {
	logf(level.ERROR, e.fields, s, a...)
}

func (e Entry) Fatal(a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, e.fields, args(len(a)), a...)
}

func (e Entry) Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	logf(level.FATAL, e.fields, s, a...)
}

func (e Entry) Panic(a ...interface{}) {
	defer panic(fmt.Sprint(a...))
	logf(level.PANIC, e.fields, args(len(a)), a...)
}

func (e Entry) Panicf(s string, a ...interface{}) {
	defer panic(fmt.Sprintf(s, a...))
	logf(level.PANIC, e.fields, s, a...)
}

func (e Entry) Log(lvl level.LEVEL, a ...interface{}) {
	logf(lvl, e.fields, args(len(a)), a...)
}

func (e Entry) Logf(lvl level.LEVEL, s string, a ...interface{}) {
	logf(lvl, e.fields, s, a...)
}

func (e Entry) Print(a ...interface{}) {
	printf(e.fields, args(len(a)), a...)
}

func (e Entry) Printf(s string, a ...interface{}) {
	printf(e.fields, s, a...)
}
