package structr

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
