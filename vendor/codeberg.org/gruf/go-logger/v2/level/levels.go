package level

// LEVEL defines a level of logging.
type LEVEL uint8

// Default levels of logging.
const (
	UNSET LEVEL = 0
	PANIC LEVEL = 1
	FATAL LEVEL = 50
	ERROR LEVEL = 100
	WARN  LEVEL = 150
	INFO  LEVEL = 200
	DEBUG LEVEL = 250
	TRACE LEVEL = 254
	ALL   LEVEL = ^LEVEL(0)
)

// CanLog returns whether an incoming log of 'lvl' can be logged against receiving level.
func (loglvl LEVEL) CanLog(lvl LEVEL) bool {
	return loglvl > lvl
}

// Levels defines a mapping of log LEVELs to formatted level strings.
type Levels [int(ALL) + 1]string

// Default returns the default set of log levels.
func Default() Levels {
	return Levels{
		TRACE: "TRACE",
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		FATAL: "FATAL",
		PANIC: "PANIC",

		// we set these just so that
		// it can be debugged when someone
		// attempts to log with ALL/UNSET
		ALL:   "{all}",
		UNSET: "{unset}",
	}
}

// Get fetches the level string for the provided value.
func (l Levels) Get(lvl LEVEL) string {
	return l[int(lvl)]
}
