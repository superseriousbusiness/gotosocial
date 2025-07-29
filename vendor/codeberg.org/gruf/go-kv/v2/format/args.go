package format

const (
	// TypeMask when set in argument flags
	// indicates that type information of
	// the passed value, and all nested types,
	// should be included in formatted output.
	TypeMask = uint64(1) << 0

	// LogfmtMask when set in argument flags
	// indicates that strings should be escaped
	// and quoted only where necessary. i.e. if
	// it contains any unsafe ASCII chars or double
	// quotes it will be quoted and escaped, if it
	// contains any spaces it will be quoted, and
	// all else will be printed as-is. This proves
	// particularly well readable in key-value types.
	LogfmtMask = uint64(1) << 1

	// NumberMask when set in argument flags
	// indicates that where possible value
	// types should be formatted as numbers,
	// i.e. byte or rune types.
	NumberMask = uint64(1) << 2

	// TextMask when set in argument flags
	// indicates that where possible value
	// types should be formatted as text,
	// i.e. []byte or []rune types.
	TextMask = uint64(1) << 3

	// QuotedTextMask when set in argument flags
	// indicates that text should always be quoted.
	QuotedTextMask = uint64(1) << 4

	// QuotedAsciiMask when set in argument flags
	// indicates that text should always be quoted,
	// and escaped as ASCII characters where needed.
	QuotedAsciiMask = uint64(1) << 5

	// NoMethodMask when set in argument flags
	// indicates that where a type supports a
	// known method, (e.g. Error() or String()),
	// this should not be used for formatting
	// instead treating as a method-less type.
	// e.g. printing the entire struct value of
	// a &url.URL{} without calling String().
	NoMethodMask = uint64(1) << 6
)

var (
	// default set of Args.
	defaultArgs = Args{
		Flags: LogfmtMask | TextMask,
		Int:   IntArgs{Base: 10},
		Uint:  IntArgs{Base: 10},
		Float: FloatArgs{Fmt: 'g', Prec: -1},
		Complex: ComplexArgs{
			Real: FloatArgs{Fmt: 'g', Prec: -1},
			Imag: FloatArgs{Fmt: 'g', Prec: -1},
		},
	}

	// zeroArgs used for checking
	// zero value Arg{} fields.
	zeroArgs Args
)

// DefaultArgs returns default
// set of formatter arguments.
func DefaultArgs() Args {
	return defaultArgs
}

// Args contains arguments
// for a call to a FormatFunc.
type Args struct {

	// Boolean argument
	// flags as bit-field.
	Flags uint64

	// Integer
	// arguments.
	// i.e. for:
	// - int
	// - int8
	// - int16
	// - int32 (treated as rune char, number with NumberMask)
	// - int64
	Int IntArgs

	// Unsigned
	// integer
	// arguments.
	// i.e. for:
	// - uint
	// - uint8 (treated as byte char, number with NumberMask)
	// - uint16
	// - uint32
	// - uint64
	Uint IntArgs

	// Float
	// arguments.
	// i.e. for:
	// - float32
	// - float64
	Float FloatArgs

	// Complex
	// arguments.
	// i.e. for:
	// - complex64
	// - complex128
	Complex ComplexArgs
}

// IntArgs provides a set of
// arguments for customizing
// integer number serialization.
type IntArgs struct {
	Base int
	Pad  int
}

// FloatArgs provides a set of
// arguments for customizing
// float number serialization.
type FloatArgs struct {
	Fmt  byte
	Prec int
}

// ComplexArgs provides a set of
// arguments for customizing complex
// number serialization, as real and
// imaginary float number parts.
type ComplexArgs struct {
	Real FloatArgs
	Imag FloatArgs
}

// WithType returns if TypeMask is set.
func (a *Args) WithType() bool {
	return a.Flags&TypeMask != 0
}

// Logfmt returns if LogfmtMask is set.
func (a *Args) Logfmt() bool {
	return a.Flags&LogfmtMask != 0
}

// AsNumber returns if NumberMask is set.
func (a *Args) AsNumber() bool {
	return a.Flags&NumberMask != 0
}

// AsText returns if TextMask is set.
func (a *Args) AsText() bool {
	return a.Flags&TextMask != 0
}

// AsQuotedText returns if QuotedTextMask is set.
func (a *Args) AsQuotedText() bool {
	return a.Flags&QuotedTextMask != 0
}

// AsQuotedASCII returns if QuotedAsciiMask is set.
func (a *Args) AsQuotedASCII() bool {
	return a.Flags&QuotedAsciiMask != 0
}

// NoMethod returns if NoMethodMask is set.
func (a *Args) NoMethod() bool {
	return a.Flags&NoMethodMask != 0
}

// SetWithType sets the TypeMask bit.
func (a *Args) SetWithType() {
	a.Flags = a.Flags | TypeMask
}

// SetLogfmt sets the LogfmtMask bit.
func (a *Args) SetLogfmt() {
	a.Flags = a.Flags | LogfmtMask
}

// SetAsNumber sets the NumberMask bit.
func (a *Args) SetAsNumber() {
	a.Flags = a.Flags | NumberMask
}

// SetAsText sets the TextMask bit.
func (a *Args) SetAsText() {
	a.Flags = a.Flags | TextMask
}

// SetAsQuotedText sets the QuotedTextMask bit.
func (a *Args) SetAsQuotedText() {
	a.Flags = a.Flags | QuotedTextMask
}

// SetAsQuotedASCII sets the QuotedAsciiMask bit.
func (a *Args) SetAsQuotedASCII() {
	a.Flags = a.Flags | QuotedAsciiMask
}

// SetNoMethod sets the NoMethodMask bit.
func (a *Args) SetNoMethod() {
	a.Flags = a.Flags | NoMethodMask
}

// UnsetWithType unsets the TypeMask bit.
func (a *Args) UnsetWithType() {
	a.Flags = a.Flags & ^TypeMask
}

// UnsetLogfmt unsets the LogfmtMask bit.
func (a *Args) UnsetLogfmt() {
	a.Flags = a.Flags & ^LogfmtMask
}

// UnsetAsNumber unsets the NumberMask bit.
func (a *Args) UnsetAsNumber() {
	a.Flags = a.Flags & ^NumberMask
}

// UnsetAsText unsets the TextMask bit.
func (a *Args) UnsetAsText() {
	a.Flags = a.Flags & ^TextMask
}

// UnsetAsQuotedText unsets the QuotedTextMask bit.
func (a *Args) UnsetAsQuotedText() {
	a.Flags = a.Flags & ^QuotedTextMask
}

// UnsetAsQuotedASCII unsets the QuotedAsciiMask bit.
func (a *Args) UnsetAsQuotedASCII() {
	a.Flags = a.Flags & ^QuotedAsciiMask
}

// UnsetNoMethod unsets the NoMethodMask bit.
func (a *Args) UnsetNoMethod() {
	a.Flags = a.Flags & ^NoMethodMask
}
