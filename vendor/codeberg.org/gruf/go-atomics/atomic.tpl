package atomics

import (
    "sync/atomic"
    "unsafe"
)

// {{ .Name }} provides user-friendly means of performing atomic operations on {{ .Type }} types.
type {{ .Name }} struct{ ptr unsafe.Pointer }

// New{{ .Name }} will return a new {{ .Name }} instance initialized with zero value.
func New{{ .Name }}() *{{ .Name }} {
    var v {{ .Type }}
    return &{{ .Name }}{
        ptr: unsafe.Pointer(&v),
    }
}

// Store will atomically store {{ .Type }} value in address contained within v.
func (v *{{ .Name }}) Store(val {{ .Type }}) {
    atomic.StorePointer(&v.ptr, unsafe.Pointer(&val))
}

// Load will atomically load {{ .Type }} value at address contained within v.
func (v *{{ .Name }}) Load() {{ .Type }} {
    return *(*{{ .Type }})(atomic.LoadPointer(&v.ptr))
}

// CAS performs a compare-and-swap for a(n) {{ .Type }} value at address contained within v.
func (v *{{ .Name }}) CAS(cmp, swp {{ .Type }}) bool {
    for {
        // Load current value at address
        ptr := atomic.LoadPointer(&v.ptr)
        cur := *(*{{ .Type }})(ptr)

        // Perform comparison against current
        if !({{ call .Compare "cur" "cmp" }}) {
            return false
        }

        // Attempt to replace pointer
        if atomic.CompareAndSwapPointer(
            &v.ptr,
            ptr,
            unsafe.Pointer(&swp),
         ) {
            return true
        }
    }
}

// Swap atomically stores new {{ .Type }} value into address contained within v, and returns previous value.
func (v *{{ .Name }}) Swap(swp {{ .Type }}) {{ .Type }} {
    ptr := unsafe.Pointer(&swp)
    ptr = atomic.SwapPointer(&v.ptr, ptr)
    return *(*{{ .Type }})(ptr)
}
