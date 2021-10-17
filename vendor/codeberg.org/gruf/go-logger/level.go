package logger

// LEVEL defines a level of logging
type LEVEL uint8

// Available levels of logging.
const (
	unset LEVEL = 255
	DEBUG LEVEL = 5
	INFO  LEVEL = 10
	WARN  LEVEL = 15
	ERROR LEVEL = 20
	FATAL LEVEL = 25
)

var unknownLevel = "unknown"

// Levels defines a mapping of log LEVELs to formatted level strings
type Levels map[LEVEL]string

// DefaultLevels returns the default set of log levels
func DefaultLevels() Levels {
	return Levels{
		DEBUG: "debug",
		INFO:  "info",
		WARN:  "warn",
		ERROR: "error",
		FATAL: "fatal",
	}
}

// LevelString fetches the appropriate level string for the provided level, or "unknown"
func (l Levels) LevelString(lvl LEVEL) string {
	str, ok := l[lvl]
	if !ok {
		return unknownLevel
	}
	return str
}
