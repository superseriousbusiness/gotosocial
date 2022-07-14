package logger

import (
	"codeberg.org/gruf/go-bitutil"
	"codeberg.org/gruf/go-logger/v2/entry"
)

// Config provides entry format configuration for a Logger.
type Config struct {
	// FlagHandler handles any necessary tasks to perform on an Entry
	// given the provided flags. Calldepth is provided for logger.FCaller.
	FlagHandler func(entry *entry.Entry, calldepth int, flags Flags)

	// Format provides formatting for Entries produced from Logger.
	Format entry.Formatter

	// Calldepth allows setting a minimum calldepth for log entries when
	// calling functions that internally set a calldepth e.g. .Print(),
	// though will not effect functions that accept a calldepth argument.
	// This is useful when wrapping Logger.
	Calldepth int
}

// defaultFlagHandler is the default implementation used when Config.FlagHandler is nil.
func defaultFlagHandler(entry *entry.Entry, calldepth int, flags Flags) {
	if flags.Time() {
		// Append timestamp
		entry.Timestamp()
	}
	if flags.Caller() {
		// Append caller information
		entry.Caller(calldepth + 1)
	}
}

// Flags provides configurable output flags for Logger.
type Flags bitutil.Flags16

// Time will return if the Time flag bit is set.
func (f Flags) Time() bool {
	return (bitutil.Flags16)(f).Get0()
}

// SetTime will set the Time flag bit.
func (f Flags) SetTime() Flags {
	return Flags((bitutil.Flags16)(f).Set0())
}

// UnsetTime will unset the Time flag bit.
func (f Flags) UnsetTime() Flags {
	return Flags((bitutil.Flags16)(f).Unset0())
}

// Caller will return if the Caller flag bit is set.
func (f Flags) Caller() bool {
	return (bitutil.Flags16)(f).Get1()
}

// SetCaller will set the Caller flag bit.
func (f Flags) SetCaller() Flags {
	return Flags((bitutil.Flags16)(f).Set1())
}

// UnsetCaller will unset the Caller flag bit.
func (f Flags) UnsetCaller() Flags {
	return Flags((bitutil.Flags16)(f).Unset1())
}
