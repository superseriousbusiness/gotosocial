package entry

import "runtime"

// caller fetches the caller info for calling function, skipping 'depth'.
func caller(depth int) (fn string, file string, line int) {
	var rpc [1]uintptr

	// Fetch pcs of callers
	n := runtime.Callers(depth+1, rpc[:])

	if n > 0 {
		// Fetch frames for determined caller pcs
		frame, _ := runtime.CallersFrames(rpc[:]).Next()
		if frame.PC != 0 {
			return frame.Function, frame.File, frame.Line
		}
	}

	// Defaults
	fn = `???`
	file = `???`
	line = -1
	return
}
