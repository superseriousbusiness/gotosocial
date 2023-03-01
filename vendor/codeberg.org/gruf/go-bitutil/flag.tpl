package bitutil

import (
    "strings"
    "unsafe"
)

{{ range $idx, $size := . }}

// Flags{{ $size.Size }} is a type-casted unsigned integer with helper
// methods for easily managing up to {{ $size.Size }} bit-flags.
type Flags{{ $size.Size }} uint{{ $size.Size }}

// Get will fetch the flag bit value at index 'bit'.
func (f Flags{{ $size.Size }}) Get(bit uint8) bool {
    mask := Flags{{ $size.Size }}(1) << bit
    return (f & mask != 0)
}

// Set will set the flag bit value at index 'bit'.
func (f Flags{{ $size.Size }}) Set(bit uint8) Flags{{ $size.Size }} {
    mask := Flags{{ $size.Size }}(1) << bit
    return f | mask
}

// Unset will unset the flag bit value at index 'bit'.
func (f Flags{{ $size.Size }}) Unset(bit uint8) Flags{{ $size.Size }} {
    mask := Flags{{ $size.Size }}(1) << bit
    return f & ^mask
}

{{ range $idx := $size.Bits }}

// Get{{ $idx }} will fetch the flag bit value at index {{ $idx }}.
func (f Flags{{ $size.Size }}) Get{{ $idx }}() bool {
    const mask = Flags{{ $size.Size }}(1) << {{ $idx }}
    return (f & mask != 0)
}

// Set{{ $idx }} will set the flag bit value at index {{ $idx }}.
func (f Flags{{ $size.Size }}) Set{{ $idx }}() Flags{{ $size.Size }} {
    const mask = Flags{{ $size.Size }}(1) << {{ $idx }}
    return f | mask
}

// Unset{{ $idx }} will unset the flag bit value at index {{ $idx }}.
func (f Flags{{ $size.Size }}) Unset{{ $idx }}() Flags{{ $size.Size }} {
    const mask = Flags{{ $size.Size }}(1) << {{ $idx }}
    return f & ^mask
}

{{ end }}

// String returns a human readable representation of Flags{{ $size.Size }}.
func (f Flags{{ $size.Size }}) String() string {
    var (
        i   int
        val bool
        buf []byte
    )

    // Make a prealloc est. based on longest-possible value
    const prealloc = 1+(len("false ")*{{ $size.Size }})-1+1
    buf = make([]byte, prealloc)

    buf[i] = '{'
    i++

    {{ range $idx := .Bits }}
    val = f.Get{{ $idx }}()
    i += copy(buf[i:], bool2str(val))
    buf[i] = ' '
    i++
    {{ end }}

    buf[i-1] = '}'
    buf = buf[:i]

    return *(*string)(unsafe.Pointer(&buf))
}

// GoString returns a more verbose human readable representation of Flags{{ $size.Size }}.
func (f Flags{{ $size.Size }})GoString() string {
    var (
        i   int
        val bool
        buf []byte
    )

    // Make a prealloc est. based on longest-possible value
    const prealloc = len("bitutil.Flags{{ $size.Size }}{")+(len("{{ sub $size.Size 1 }}=false ")*{{ $size.Size }})-1+1
    buf = make([]byte, prealloc)

    i += copy(buf[i:], "bitutil.Flags{{ $size.Size }}{")

    {{ range $idx := .Bits }}
    val = f.Get{{ $idx }}()
    i += copy(buf[i:], "{{ $idx }}=")
    i += copy(buf[i:], bool2str(val))
    buf[i] = ' '
    i++
    {{ end }}
 
    buf[i-1] = '}'
    buf = buf[:i]

    return *(*string)(unsafe.Pointer(&buf))
}

{{ end }}

func bool2str(b bool) string {
    if b {
        return "true"
    }
    return "false"
}