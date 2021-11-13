package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"codeberg.org/gruf/go-bytes"
)

type Logger struct {
	// Hooks defines a list of hooks which are called before an entry
	// is written. This should NOT be modified while the Logger is in use
	Hooks []Hook

	// Level is the current log LEVEL, entries at level below the
	// currently set level will not be output. This should NOT
	// be modified while the Logger is in use
	Level LEVEL

	// Timestamp defines whether to automatically append timestamps
	// to entries written via Logger convience methods and specifically
	// Entry.TimestampIf(). This should NOT be modified while Logger in use
	Timestamp bool

	// Format is the log entry LogFormat to use. This should NOT
	// be modified while the Logger is in use
	Format LogFormat

	// BufferSize is the Entry buffer size to use when allocating
	// new Entry objects. This should be modified atomically
	BufSize int64

	// Output is the log's output writer. This should NOT be
	// modified while the Logger is in use
	Output io.Writer

	// entry pool
	pool sync.Pool
}

// New returns a new Logger instance with defaults
func New(out io.Writer) *Logger {
	return NewWith(0 /* all */, true, NewLogFmt(false), 512, out)
}

// NewWith returns a new Logger instance with supplied configuration
func NewWith(lvl LEVEL, timestamp bool, fmt LogFormat, bufsize int64, out io.Writer) *Logger {
	// Create new logger object
	log := &Logger{
		Level:     lvl,
		Timestamp: timestamp,
		Format:    fmt,
		BufSize:   bufsize,
		Output:    out,
		pool:      sync.Pool{},
	}

	// Ensure clock running
	startClock()

	// Set-up logger Entry pool
	log.pool.New = func() interface{} {
		return &Entry{
			lvl: unset,
			buf: &bytes.Buffer{B: make([]byte, 0, atomic.LoadInt64(&log.BufSize))},
			log: log,
		}
	}

	return log
}

// Entry returns a new Entry from the Logger's pool with background context
func (l *Logger) Entry() *Entry {
	return l.pool.Get().(*Entry).WithContext(context.Background())
}

// Debug prints the provided arguments with the debug prefix
func (l *Logger) Debug(a ...interface{}) {
	l.Entry().TimestampIf().Level(DEBUG).Hooks().Msg(a...)
}

// Debugf prints the provided format string and arguments with the debug prefix
func (l *Logger) Debugf(s string, a ...interface{}) {
	l.Entry().TimestampIf().Level(DEBUG).Hooks().Msgf(s, a...)
}

// Info prints the provided arguments with the info prefix
func (l *Logger) Info(a ...interface{}) {
	l.Entry().TimestampIf().Level(INFO).Hooks().Msg(a...)
}

// Infof prints the provided format string and arguments with the info prefix
func (l *Logger) Infof(s string, a ...interface{}) {
	l.Entry().TimestampIf().Level(INFO).Hooks().Msgf(s, a...)
}

// Warn prints the provided arguments with the warn prefix
func (l *Logger) Warn(a ...interface{}) {
	l.Entry().TimestampIf().Level(WARN).Hooks().Msg(a...)
}

// Warnf prints the provided format string and arguments with the warn prefix
func (l *Logger) Warnf(s string, a ...interface{}) {
	l.Entry().TimestampIf().Level(WARN).Hooks().Msgf(s, a...)
}

// Error prints the provided arguments with the error prefix
func (l *Logger) Error(a ...interface{}) {
	l.Entry().TimestampIf().Level(ERROR).Hooks().Msg(a...)
}

// Errorf prints the provided format string and arguments with the error prefix
func (l *Logger) Errorf(s string, a ...interface{}) {
	l.Entry().TimestampIf().Level(ERROR).Hooks().Msgf(s, a...)
}

// Fatal prints provided arguments with the fatal prefix before exiting the program
// with os.Exit(1)
func (l *Logger) Fatal(a ...interface{}) {
	defer os.Exit(1)
	l.Entry().TimestampIf().Level(FATAL).Hooks().Msg(a...)
}

// Fatalf prints provided the provided format string and arguments with the fatal prefix
// before exiting the program with os.Exit(1)
func (l *Logger) Fatalf(s string, a ...interface{}) {
	defer os.Exit(1)
	l.Entry().TimestampIf().Level(FATAL).Hooks().Msgf(s, a...)
}

// Log prints the provided arguments with the supplied log level
func (l *Logger) Log(lvl LEVEL, a ...interface{}) {
	l.Entry().TimestampIf().Hooks().Msg(a...)
}

// Logf prints the provided format string and arguments with the supplied log level
func (l *Logger) Logf(lvl LEVEL, s string, a ...interface{}) {
	l.Entry().TimestampIf().Hooks().Msgf(s, a...)
}

// Print simply prints provided arguments
func (l *Logger) Print(a ...interface{}) {
	l.Entry().Hooks().Msg(a...)
}

// Printf simply prints provided the provided format string and arguments
func (l *Logger) Printf(s string, a ...interface{}) {
	l.Entry().Hooks().Msgf(s, a...)
}
