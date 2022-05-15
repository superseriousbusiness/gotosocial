//go:build debugenv
// +build debugenv

package debug

import "os"

// check if debug env variable is set
var debug = (os.Getenv("DEBUG") != "")
