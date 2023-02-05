package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestIDKey is a string to use as a map key, for example a logger field
const RequestIDKey = "requestID"

// RequestID returns a gin middleware which adds a unique ID for each request
// to the context. It currently directly wraps the upstream library but this
// makes it easier to set any custom options later on.
func RequestID() gin.HandlerFunc {
	return requestid.New()
}
