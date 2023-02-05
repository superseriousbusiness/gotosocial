package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var headerXRequestID string

// Config defines the config for RequestID middleware
type config struct {
	// Generator defines a function to generate an ID.
	// Optional. Default: func() string {
	//   return uuid.New().String()
	// }
	generator Generator
	headerKey HeaderStrKey
	handler   Handler
}

// New initializes the RequestID middleware.
func New(opts ...Option) gin.HandlerFunc {
	cfg := &config{
		generator: func() string {
			return uuid.New().String()
		},
		headerKey: "X-Request-ID",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	headerXRequestID = string(cfg.headerKey)

	return func(c *gin.Context) {
		// Get id from request
		rid := c.GetHeader(headerXRequestID)
		if rid == "" {
			rid = cfg.generator()
			c.Request.Header.Add(headerXRequestID, rid)
		}
		if cfg.handler != nil {
			cfg.handler(c, rid)
		}
		// Set the id to ensure that the requestid is in the response
		c.Header(headerXRequestID, rid)
		c.Next()
	}
}

// Get returns the request identifier
func Get(c *gin.Context) string {
	return c.Writer.Header().Get(headerXRequestID)
}
