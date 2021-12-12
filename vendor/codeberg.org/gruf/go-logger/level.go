package logger

// LEVEL defines a level of logging
type LEVEL uint8

// Available levels of logging.
const (
	unset LEVEL = ^LEVEL(0)
	DEBUG LEVEL = 5
	INFO  LEVEL = 10
	WARN  LEVEL = 15
	ERROR LEVEL = 20
	FATAL LEVEL = 25
)

var unknownLevel = "unknown"

// Levels defines a mapping of log LEVELs to formatted level strings
type Levels [^LEVEL(0)]string

// DefaultLevels returns the default set of log levels
func DefaultLevels() Levels {
	return Levels{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		FATAL: "FATAL",
	}
}

// Get fetches the level string for the provided value, or "unknown"
func (l Levels) Get(lvl LEVEL) string {
	if str := l[int(lvl)]; str != "" {
		return str
	}
	return unknownLevel
}
