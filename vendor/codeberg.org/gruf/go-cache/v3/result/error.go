package result

// IsConflictErr returns whether error is due to key conflict.
func IsConflictErr(err error) bool {
	_, ok := err.(ConflictError)
	return ok
}

// ConflictError is returned on cache key conflict.
type ConflictError struct {
	Key string
}

// Error returns the message for this key conflict error.
func (c ConflictError) Error() string {
	return "cache conflict for key \"" + c.Key + "\""
}
