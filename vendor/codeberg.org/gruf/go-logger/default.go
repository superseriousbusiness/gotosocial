package logger

import (
	"os"
	"sync"
)

var (
	instance     *Logger
	instanceOnce = sync.Once{}
)

// Default returns the default Logger instance.
func Default() *Logger {
	instanceOnce.Do(func() { instance = New(os.Stdout) })
	return instance
}

// Debug prints the provided arguments with the debug prefix to the global Logger instance.
func Debug(a ...interface{}) {
	Default().Debug(a...)
}

// Debugf prints the provided format string and arguments with the debug prefix to the global Logger instance.
func Debugf(s string, a ...interface{}) {
	Default().Debugf(s, a...)
}

// Info prints the provided arguments with the info prefix to the global Logger instance.
func Info(a ...interface{}) {
	Default().Info(a...)
}

// Infof prints the provided format string and arguments with the info prefix to the global Logger instance.
func Infof(s string, a ...interface{}) {
	Default().Infof(s, a...)
}

// Warn prints the provided arguments with the warn prefix to the global Logger instance.
func Warn(a ...interface{}) {
	Default().Warn(a...)
}

// Warnf prints the provided format string and arguments with the warn prefix to the global Logger instance.
func Warnf(s string, a ...interface{}) {
	Default().Warnf(s, a...)
}

// Error prints the provided arguments with the error prefix to the global Logger instance.
func Error(a ...interface{}) {
	Default().Error(a...)
}

// Errorf prints the provided format string and arguments with the error prefix to the global Logger instance.
func Errorf(s string, a ...interface{}) {
	Default().Errorf(s, a...)
}

// Fatal prints the provided arguments with the fatal prefix to the global Logger instance before exiting the program with os.Exit(1).
func Fatal(a ...interface{}) {
	Default().Fatal(a...)
}

// Fatalf prints the provided format string and arguments with the fatal prefix to the global Logger instance before exiting the program with os.Exit(1).
func Fatalf(s string, a ...interface{}) {
	Default().Fatalf(s, a...)
}

// Log prints the provided arguments with the supplied log level to the global Logger instance.
func Log(lvl LEVEL, a ...interface{}) {
	Default().Log(lvl, a...)
}

// Logf prints the provided format string and arguments with the supplied log level to the global Logger instance.
func Logf(lvl LEVEL, s string, a ...interface{}) {
	Default().Logf(lvl, s, a...)
}

// LogFields prints the provided fields formatted as key-value pairs at the supplied log level to the global Logger instance.
func LogFields(lvl LEVEL, fields map[string]interface{}) {
	Default().LogFields(lvl, fields)
}

// LogValues prints the provided values formatted as-so at the supplied log level to the global Logger instance.
func LogValues(lvl LEVEL, a ...interface{}) {
	Default().LogValues(lvl, a...)
}

// Print simply prints provided arguments to the global Logger instance.
func Print(a ...interface{}) {
	Default().Print(a...)
}

// Printf simply prints provided the provided format string and arguments to the global Logger instance.
func Printf(s string, a ...interface{}) {
	Default().Printf(s, a...)
}

// PrintFields prints the provided fields formatted as key-value pairs to the global Logger instance.
func PrintFields(fields map[string]interface{}) {
	Default().PrintFields(fields)
}

// PrintValues prints the provided values formatted as-so to the global Logger instance.
func PrintValues(a ...interface{}) {
	Default().PrintValues(a...)
}
