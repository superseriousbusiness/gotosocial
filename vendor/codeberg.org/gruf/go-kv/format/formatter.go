package format

import (
	"strconv"
	"strings"

	"codeberg.org/gruf/go-byteutil"
)

// Formatter allows configuring value and string formatting.
type Formatter struct {
	// MaxDepth specifies the max depth of fields the formatter will iterate.
	// Once max depth is reached, value will simply be formatted as "...".
	// e.g.
	//
	// MaxDepth=1
	// type A struct{
	//     Nested B
	// }
	// type B struct{
	//     Nested C
	// }
	// type C struct{
	//     Field string
	// }
	//
	// Append(&buf, A{}) => {Nested={Nested={Field=...}}}
	MaxDepth uint8
}

// Append will append formatted form of supplied values into 'buf'.
func (f *Formatter) Append(buf *byteutil.Buffer, v ...interface{}) {
	fmt := format{Buffer: buf, Config: f}
	for i := 0; i < len(v); i++ {
		fmt.AppendInterfaceOrReflect(v[i])
		fmt.Buffer.B = append(fmt.Buffer.B, ' ')
	}
	if len(v) > 0 {
		fmt.Buffer.Truncate(1)
	}
}

// Appendf will append the formatted string with supplied values into 'buf'.
// Supported format directives:
// - '{}'   => format supplied arg, in place
// - '{0}'  => format arg at index 0 of supplied, in place
// - '{:?}' => format supplied arg verbosely, in place
// - '{:k}' => format supplied arg as key, in place
// - '{:v}' => format supplied arg as value, in place
//
// To escape either of '{}' simply append an additional brace e.g.
// - '{{'     => '{'
// - '}}'     => '}'
// - '{{}}'   => '{}'
// - '{{:?}}' => '{:?}'
//
// More formatting directives might be included in the future.
func (f *Formatter) Appendf(buf *byteutil.Buffer, s string, a ...interface{}) {
	const (
		// ground state
		modeNone = uint8(0)

		// prev reached '{'
		modeOpen = uint8(1)

		// prev reached '}'
		modeClose = uint8(2)

		// parsing directive index
		modeIdx = uint8(3)

		// parsing directive operands
		modeOp = uint8(4)
	)

	var (
		// mode is current parsing mode
		mode uint8

		// arg is the current arg index
		arg int

		// carg is current directive-set arg index
		carg int

		// last is the trailing cursor to see slice windows
		last int

		// idx is the current index in 's'
		idx int

		// fmt is the base argument formatter
		fmt = format{
			Config: f,
			Buffer: buf,
		}

		// NOTE: these functions are defined here as function
		// locals as it turned out to be better for performance
		// doing it this way, than encapsulating their logic in
		// some kind of parsing structure. Maybe if the parser
		// was pooled along with the buffers it might work out
		// better, but then it makes more internal functions i.e.
		// .Append() .Appendf() less accessible outside package.
		//
		// Currently, passing '-gcflags "-l=4"' causes a not
		// insignificant decrease in ns/op, which is likely due
		// to more aggressive function inlining, which this
		// function can obviously stand to benefit from :)

		// Str returns current string window slice, and updates
		// the trailing cursor 'last' to current 'idx'
		Str = func() string {
			str := s[last:idx]
			last = idx
			return str
		}

		// MoveUp moves the trailing cursor 'last' just past 'idx'
		MoveUp = func() {
			last = idx + 1
		}

		// MoveUpTo moves the trailing cursor 'last' either up to
		// closest '}', or current 'idx', whichever is furthest.
		// NOTE: by calling bytealg.IndexByteString() directly (and
		// not the strconv pass-through, we shave-off complexity
		// which allows this function to be inlined).
		MoveUpTo = func() {
			i := strings.IndexByte(s[idx:], '}')
			if i >= 0 {
				idx += i
			}
			MoveUp()
		}

		// ParseIndex parses an integer from the current string
		// window, updating 'last' to 'idx'. The string window
		// is ASSUMED to contain only valid ASCII numbers. This
		// only returns false if number exceeds platform int size
		ParseIndex = func() bool {
			var str string

			// Get current window
			if str = Str(); len(str) < 1 {
				return true
			}

			// Index HAS to fit within platform integer size
			if !(strconv.IntSize == 32 && (0 < len(s) && len(s) < 10)) ||
				!(strconv.IntSize == 64 && (0 < len(s) && len(s) < 19)) {
				return false
			}

			carg = 0

			// Build integer from string
			for i := 0; i < len(str); i++ {
				carg = carg*10 + int(str[i]-'0')
			}

			return true
		}

		// ValidOp checks that for current ending idx, that a valid
		// operand was achieved -- only 0, 1 are valid numbers.
		ValidOp = func() bool {
			diff := (idx - last)
			last = idx
			return diff < 2
		}

		// AppendArg will take either the directive-set, or
		// iterated arg index, check within bounds of 'a' and
		// append the that argument formatted to the buffer.
		// On failure, it will append an error string
		AppendArg = func() {
			// Look for idx
			if carg < 0 {
				carg = arg
			}

			// Incr idx
			arg++

			if carg < len(a) {
				// Append formatted argument value
				fmt.AppendInterfaceOrReflect(a[carg])
			} else {
				// No argument found for index
				fmt.Buffer.B = append(fmt.Buffer.B, `!{MISSING_ARG}`...)
			}
		}

		// Reset will reset the mode to ground, the flags
		// to empty and parsed 'carg' to  empty
		Reset = func() {
			mode = modeNone
			fmt.CurDepth = 0
			fmt.Flags = 0
			fmt.VType = ""
			fmt.Derefs = 0
			carg = -1
		}
	)

	for idx = 0; idx < len(s); idx++ {
		// Get next char
		c := s[idx]

		switch mode {
		// Ground mode
		case modeNone:
			switch c {
			case '{':
				// Enter open mode
				fmt.Buffer.B = append(fmt.Buffer.B, Str()...)
				mode = modeOpen
				MoveUp()
			case '}':
				// Enter close mode
				fmt.Buffer.B = append(fmt.Buffer.B, Str()...)
				mode = modeClose
				MoveUp()
			}

		// Encountered open '{'
		case modeOpen:
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				// Starting index
				mode = modeIdx
				MoveUp()
			case '{':
				// Escaped bracket
				fmt.Buffer.B = append(fmt.Buffer.B, '{')
				mode = modeNone
				MoveUp()
			case '}':
				// Format arg
				AppendArg()
				Reset()
				MoveUp()
			case ':':
				// Starting operands
				mode = modeOp
				MoveUp()
			default:
				// Bad char, missing a close
				fmt.Buffer.B = append(fmt.Buffer.B, `!{MISSING_CLOSE}`...)
				mode = modeNone
				MoveUpTo()
			}

		// Encountered close '}'
		case modeClose:
			switch c {
			case '}':
				// Escaped close bracket
				fmt.Buffer.B = append(fmt.Buffer.B, '}')
				mode = modeNone
				MoveUp()
			default:
				// Missing an open bracket
				fmt.Buffer.B = append(fmt.Buffer.B, `!{MISSING_OPEN}`...)
				mode = modeNone
				MoveUp()
			}

		// Preparing index
		case modeIdx:
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			case ':':
				if !ParseIndex() {
					// Unable to parse an integer
					fmt.Buffer.B = append(fmt.Buffer.B, `!{BAD_INDEX}`...)
					mode = modeNone
					MoveUpTo()
				} else {
					// Starting operands
					mode = modeOp
					MoveUp()
				}
			case '}':
				if !ParseIndex() {
					// Unable to parse an integer
					fmt.Buffer.B = append(fmt.Buffer.B, `!{BAD_INDEX}`...)
				} else {
					// Format arg
					AppendArg()
				}
				Reset()
				MoveUp()
			default:
				// Not a valid index character
				fmt.Buffer.B = append(fmt.Buffer.B, `!{BAD_INDEX}`...)
				mode = modeNone
				MoveUpTo()
			}

		// Preparing operands
		case modeOp:
			switch c {
			case 'k':
				fmt.Flags |= IsKeyBit
			case 'v':
				fmt.Flags |= IsValBit
			case '?':
				fmt.Flags |= VboseBit
			case '}':
				if !ValidOp() {
					// Bad operands parsed
					fmt.Buffer.B = append(fmt.Buffer.B, `!{BAD_OPERAND}`...)
				} else {
					// Format arg
					AppendArg()
				}
				Reset()
				MoveUp()
			default:
				// Not a valid operand char
				fmt.Buffer.B = append(fmt.Buffer.B, `!{BAD_OPERAND}`...)
				Reset()
				MoveUpTo()
			}
		}
	}

	// Append any remaining
	fmt.Buffer.B = append(fmt.Buffer.B, s[last:]...)
}

// formatter is the default formatter instance.
var formatter = Formatter{
	MaxDepth: 10,
}

// Append will append formatted form of supplied values into 'buf' using default formatter.
// See Formatter.Append() for more documentation.
func Append(buf *byteutil.Buffer, v ...interface{}) {
	formatter.Append(buf, v...)
}

// Appendf will append the formatted string with supplied values into 'buf' using default formatter.
// See Formatter.Appendf() for more documentation.
func Appendf(buf *byteutil.Buffer, s string, a ...interface{}) {
	formatter.Appendf(buf, s, a...)
}
