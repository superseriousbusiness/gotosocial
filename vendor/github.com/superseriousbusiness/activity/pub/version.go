package pub

import (
	"fmt"
)

const (
	// Version string, used in the User-Agent
	version = "v1.0.0"
)

// goFedUserAgent returns the user agent string for the go-fed library.
func goFedUserAgent() string {
	return fmt.Sprintf("(go-fed/activity %s)", version)
}
