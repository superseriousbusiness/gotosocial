package debug

// DEBUG returns whether debugging is enabled.
func DEBUG() bool {
	return debug
}

// Run will only call fn if DEBUG is enabled.
func Run(fn func()) {
	if debug {
		fn()
	}
}
