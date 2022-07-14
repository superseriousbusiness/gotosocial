package logger

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"syscall"

	"codeberg.org/gruf/go-bitutil"
	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/entry"
	"codeberg.org/gruf/go-logger/v2/level"
)

// Logger is an object that provides logging to an io.Writer output
// with levels and many formatting options. Each log entry is written
// to the output by a singular .Write() call.
type Logger struct {
	output io.Writer  // output is the actual log output
	mutex  sync.Mutex // mutex protects concurrent writes
	config Config     // config contains the current logging config
	packd  uint32     // packd contains the log level and flags values
}

// New returns a new Logger instance with defaults.
func New(out io.Writer) *Logger {
	return NewWith(out, Config{}, 0, Flags(0).SetTime())
}

// NewWith returns a new Logger instance with given configuration, logging level and log flags.
func NewWith(out io.Writer, cfg Config, lvl level.LEVEL, flags Flags) *Logger {
	if cfg.FlagHandler == nil {
		// Use default flag handler if nil
		cfg.FlagHandler = defaultFlagHandler
	}

	// NOTE: we allow passing a nil
	// cfg.Format as the log entry.Entry{}
	// uses a default when Formatter nil.

	// Create new logger object
	log := &Logger{
		output: out,
		config: cfg,
	}

	// Set level and flags
	log.SetLevel(lvl)
	log.SetFlags(flags)

	return log
}

func (l *Logger) Trace(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.TRACE, a...)
}

func (l *Logger) Tracef(s string, a ...interface{}) {
	l.Logf(l.config.Calldepth+2, level.TRACE, s, a...)
}

func (l *Logger) TraceKVs(fields ...kv.Field) {
	l.LogKVs(l.config.Calldepth+2, level.TRACE, fields...)
}

func (l *Logger) Debug(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.DEBUG, a...)
}

func (l *Logger) Debugf(s string, a ...interface{}) {
	l.Logf(l.config.Calldepth+2, level.DEBUG, s, a...)
}

func (l *Logger) DebugKVs(fields ...kv.Field) {
	l.LogKVs(l.config.Calldepth+2, level.DEBUG, fields...)
}

func (l *Logger) Info(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.INFO, a...)
}

func (l *Logger) Infof(s string, a ...interface{}) {
	l.Logf(l.config.Calldepth+2, level.INFO, s, a...)
}

func (l *Logger) InfoKVs(fields ...kv.Field) {
	l.LogKVs(l.config.Calldepth+2, level.INFO, fields...)
}

func (l *Logger) Warn(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.WARN, a...)
}

func (l *Logger) Warnf(s string, a ...interface{}) {
	l.Logf(l.config.Calldepth+2, level.WARN, s, a...)
}

func (l *Logger) WarnKVs(fields ...kv.Field) {
	l.LogKVs(l.config.Calldepth+2, level.WARN, fields...)
}

func (l *Logger) Error(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.ERROR, a...)
}

func (l *Logger) Errorf(s string, a ...interface{}) {
	l.Logf(l.config.Calldepth+2, level.ERROR, s, a...)
}

func (l *Logger) ErrorKVs(fields ...kv.Field) {
	l.LogKVs(l.config.Calldepth+2, level.ERROR, fields...)
}

func (l *Logger) Fatal(a ...interface{}) {
	l.Log(l.config.Calldepth+2, level.FATAL, a...)
	syscall.Exit(1)
}

func (l *Logger) Fatalf(s string, a ...interface{}) {
	defer syscall.Exit(1)
	l.Logf(l.config.Calldepth+2, level.FATAL, s, a...)
}

func (l *Logger) FatalKVs(fields ...kv.Field) {
	defer syscall.Exit(1)
	l.LogKVs(l.config.Calldepth+2, level.FATAL, fields...)
}

func (l *Logger) Panic(a ...interface{}) {
	str := fmt.Sprint(a...)
	defer panic(str)
	l.Log(l.config.Calldepth+2, level.PANIC, str)
}

func (l *Logger) Panicf(s string, a ...interface{}) {
	str := fmt.Sprintf(s, a...)
	defer panic(str)
	l.Log(l.config.Calldepth+2, level.PANIC, str)
}

func (l *Logger) PanicKVs(fields ...kv.Field) {
	str := kv.Fields(fields).String()
	defer panic(str)
	l.Log(l.config.Calldepth+2, level.PANIC, str)
}

// Print writes provided args to the log output.
func (l *Logger) Print(a ...interface{}) {
	e := l.Entry(l.config.Calldepth + 2)
	e.Msg(a...)
	l.Write(e)
}

// Printf writes provided format string and args to the log output.
func (l *Logger) Printf(s string, a ...interface{}) {
	e := l.Entry(l.config.Calldepth + 2)
	e.Msgf(s, a...)
	l.Write(e)
}

// PrintKVs writes provided key-value fields to the log output.
func (l *Logger) PrintKVs(fields ...kv.Field) {
	e := l.Entry(l.config.Calldepth + 2)
	e.Fields(fields...)
	l.Write(e)
}

// Log writes provided args to the log output at provided log level. Calldepth determines
// number of frames to skip when .FCaller is enabled and calculating stack frames.
func (l *Logger) Log(calldepth int, lvl level.LEVEL, a ...interface{}) {
	loglvl, flags := l.unpack()
	if !loglvl.CanLog(lvl) {
		return
	}
	entry := l.NewEntry()
	l.config.FlagHandler(entry, calldepth+1, flags)
	entry.WithLevel(lvl)
	entry.Msg(a...)
	l.Write(entry)
}

// Logf writes provided format string and args to the log output. Calldepth determines
// number of frames to skip when .FCaller is enabled and calculating stack frames.
func (l *Logger) Logf(calldepth int, lvl level.LEVEL, s string, a ...interface{}) {
	loglvl, flags := l.unpack()
	if !loglvl.CanLog(lvl) {
		return
	}
	entry := l.NewEntry()
	l.config.FlagHandler(entry, calldepth+1, flags)
	entry.WithLevel(lvl)
	entry.Msgf(s, a...)
	l.Write(entry)
}

// LogKVs writes provided key-value fields to the log output. Calldepth determines
// number of frames to skip when .FCaller is enabled and calculating stack frames.
func (l *Logger) LogKVs(calldepth int, lvl level.LEVEL, fields ...kv.Field) {
	loglvl, flags := l.unpack()
	if !loglvl.CanLog(lvl) {
		return
	}
	entry := l.NewEntry()
	l.config.FlagHandler(entry, calldepth+1, flags)
	entry.WithLevel(lvl)
	entry.Fields(fields...)
	l.Write(entry)
}

// Level will return the currently set logger value.
func (l *Logger) Level() level.LEVEL {
	lvl, _ := l.unpack()
	return lvl
}

// SetLevel will set the logger level to the given value.
func (l *Logger) SetLevel(lvl level.LEVEL) {
	_, flags := l.unpack()
	l.pack(lvl, flags)
}

// Flags will return the currently set logger flags.
func (l *Logger) Flags() Flags {
	_, flags := l.unpack()
	return flags
}

// SetFlags will set the logger flags to the given value.
func (l *Logger) SetFlags(flags Flags) {
	lvl, _ := l.unpack()
	l.pack(lvl, flags)
}

// Writer returns the current output writer.
func (l *Logger) Writer() io.Writer {
	l.mutex.Lock()
	out := l.output
	l.mutex.Unlock()
	return out
}

// SetOutput will update the log output to given writer.
func (l *Logger) SetOutput(out io.Writer) {
	l.mutex.Lock()
	l.output = out
	l.mutex.Unlock()
}

// Entry will acquire a log entry from memory pool and append
// any information required by log flags using Config.FlagHandler().
// This is valid until passed to Logger.Write().
func (l *Logger) Entry(calldepth int) *entry.Entry {
	entry := l.NewEntry()
	l.config.FlagHandler(entry, calldepth+1, l.Flags())
	return entry
}

// NewEntry will acquire a log entry from memory pool and
// pass directly to the caller without calling Config.FlagHandler().
// This is valid until passed to Logger.Write().
func (l *Logger) NewEntry() *entry.Entry {
	return getEntry(l.config.Format)
}

// Write will write the contents of log entry to the output
// writer, reset it, then release to memory pool. The entry
// is NOT safe to be used after passing to this function.
func (l *Logger) Write(entry *entry.Entry) {
	// Drop any trailing space
	if buf := entry.Buffer(); // nocollapse
	len(buf.B) > 0 && buf.B[len(buf.B)-1] == ' ' {
		buf.B = buf.B[:len(buf.B)-1]
	}

	// Ensure a final new-line
	if buf := entry.Buffer(); // nocollapse
	len(buf.B) > 0 && buf.B[len(buf.B)-1] != '\n' {
		buf.B = append(buf.B, '\n')
	}

	// Acquire mutex
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// write the entry data to output
	_, _ = l.output.Write(entry.Bytes())

	// Release Entry
	putEntry(entry)
}

// pack will atomically store logger level and flags.
func (l *Logger) pack(lvl level.LEVEL, flags Flags) {
	packed := bitutil.PackUint16s(uint16(lvl), uint16(flags))
	atomic.StoreUint32(&l.packd, packed)
}

// unpack will atomically load logger level and flags.
func (l *Logger) unpack() (level.LEVEL, Flags) {
	packed := atomic.LoadUint32(&l.packd)
	lvl, flags := bitutil.UnpackUint16s(packed)
	return level.LEVEL(lvl), Flags(flags)
}
