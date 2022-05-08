package bitutil_test

import (
    "strings"
    "testing"

    "codeberg.org/gruf/go-bytes"
)

{{ range $idx, $size := . }}

func TestFlags{{ $size.Size }}Get(t *testing.T) {
    var mask, flags bitutil.Flags{{ $size.Size }}

    {{ range $idx := $size.Bits }}

    mask = bitutil.Flags{{ $size.Size }}(1) << {{ $idx }}

    flags = 0

    flags |= mask
    if !flags.Get({{ $idx }}) {
        t.Error("failed .Get() set Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    flags = ^bitutil.Flags{{ $size.Size }}(0)

    flags &= ^mask
    if flags.Get({{ $idx }}) {
        t.Error("failed .Get() unset Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    flags = 0

    flags |= mask
    if !flags.Get{{ $idx }}() {
        t.Error("failed .Get{{ $idx }}() set Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    flags = ^bitutil.Flags{{ $size.Size }}(0)

    flags &= ^mask
    if flags.Get{{ $idx }}() {
        t.Error("failed .Get{{ $idx }}() unset Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    {{ end }}
}

func TestFlags{{ $size.Size }}Set(t *testing.T) {
    var mask, flags bitutil.Flags{{ $size.Size }}

    {{ range $idx := $size.Bits }}

    mask = bitutil.Flags{{ $size.Size }}(1) << {{ $idx }}

    flags = 0

    flags = flags.Set({{ $idx }})
    if flags & mask == 0 {
        t.Error("failed .Set() Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    flags = 0

    flags = flags.Set{{ $idx }}()
    if flags & mask == 0 {
        t.Error("failed .Set{{ $idx }}() Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    {{ end }}
}

func TestFlags{{ $size.Size }}Unset(t *testing.T) {
    var mask, flags bitutil.Flags{{ $size.Size }}

    {{ range $idx := $size.Bits }}

    mask = bitutil.Flags{{ $size.Size }}(1) << {{ $idx }}

    flags = ^bitutil.Flags{{ $size.Size }}(0)

    flags = flags.Unset({{ $idx }})
    if flags & mask != 0 {
        t.Error("failed .Unset() Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    flags = ^bitutil.Flags{{ $size.Size }}(0)

    flags = flags.Unset{{ $idx }}()
    if flags & mask != 0 {
        t.Error("failed .Unset{{ $idx }}() Flags{{ $size.Size }} bit at index {{ $idx }}")
    }

    {{ end }}
}

{{ end }}