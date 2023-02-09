//go:build debugenv
// +build debugenv

package debug

import "os"

// DEBUG returns whether debugging is enabled.
var DEBUG = (os.Getenv("DEBUG") != "")
