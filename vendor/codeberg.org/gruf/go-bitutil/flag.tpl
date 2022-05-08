package bitutil

import (
    "strings"

    "codeberg.org/gruf/go-bytes"
)

{{ range $idx, $size := . }}

// Flags{{ $size.Size }} is a type-casted unsigned integer with helper
// methods for easily managing up to {{ $size.Size }} bit flags.
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
    var val bool
    var buf bytes.Buffer

    buf.WriteByte('{')
    {{ range $idx := .Bits }}
    val = f.Get{{ $idx }}()
    buf.WriteString(bool2str(val) + " ")
    {{ end }}
    buf.Truncate(1)
    buf.WriteByte('}')

    return buf.String()
}

// GoString returns a more verbose human readable representation of Flags{{ $size.Size }}.
func (f Flags{{ $size.Size }})GoString() string {
    var val bool
    var buf bytes.Buffer

    buf.WriteString("bitutil.Flags{{ $size.Size }}{")
    {{ range $idx := .Bits }}
    val = f.Get{{ $idx }}()
    buf.WriteString("{{ $idx }}="+bool2str(val)+" ")
    {{ end }}
    buf.Truncate(1)
    buf.WriteByte('}')

    return buf.String()
}

{{ end }}

func bool2str(b bool) string {
    if b {
        return "true"
    }
    return "false"
}