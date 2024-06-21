package structr

import "unsafe"

// once only executes 'fn' once.
func once(fn func()) func() {
	var once int32
	return func() {
		if once != 0 {
			return
		}
		once = 1
		fn()
	}
}

// eface_data returns the data ptr from an empty interface.
func eface_data(a any) unsafe.Pointer {
	type eface struct{ _, data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&a)).data
}
