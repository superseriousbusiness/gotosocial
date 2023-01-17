package errors

import (
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// Callers is a stacktrace of caller PCs.
type Callers []uintptr

// GetCallers returns a Callers slice of PCs, of at most 'depth'.
func GetCallers(skip int, depth int) Callers {
	rpc := make([]uintptr, depth)
	n := runtime.Callers(skip+1, rpc)
	return Callers(rpc[0:n])
}

// Frames fetches runtime frames for a slice of caller PCs.
func (f Callers) Frames() []runtime.Frame {
	// Allocate expected frames slice
	frames := make([]runtime.Frame, 0, len(f))

	// Get frames iterator for PCs
	iter := runtime.CallersFrames(f)

	for {
		// Get next frame in iter
		frame, ok := iter.Next()
		if !ok {
			break
		}

		// Append to frames slice
		frames = append(frames, frame)
	}

	return frames
}

// MarshalJSON implements json.Marshaler to provide an easy, simple default.
func (f Callers) MarshalJSON() ([]byte, error) {
	// JSON-able frame type
	type jsonFrame struct {
		Func string `json:"func"`
		File string `json:"file"`
		Line int    `json:"line"`
	}

	// Convert to frames
	frames := f.Frames()

	// Allocate expected size jsonFrame slice
	jsonFrames := make([]jsonFrame, 0, len(f))

	for i := 0; i < len(frames); i++ {
		frame := frames[i]

		// Convert each to jsonFrame object
		jsonFrames = append(jsonFrames, jsonFrame{
			Func: funcname(frame.Function),
			File: frame.File,
			Line: frame.Line,
		})
	}

	// marshal converted frames
	return json.Marshal(frames)
}

// String will return a simple string representation of receiving Callers slice.
func (f Callers) String() string {
	// Guess-timate to reduce allocs
	buf := make([]byte, 0, 64*len(f))

	// Convert to frames
	frames := f.Frames()

	for i := 0; i < len(frames); i++ {
		frame := frames[i]

		// Append formatted caller info
		fn := funcname(frame.Function)
		buf = append(buf, fn+"()\n\t"+frame.File+":"...)
		buf = strconv.AppendInt(buf, int64(frame.Line), 10)
		buf = append(buf, '\n')
	}

	return *(*string)(unsafe.Pointer(&buf))
}

// funcname splits a function name with pkg from its path prefix.
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	return name[i+1:]
}
