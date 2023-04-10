package gtscontext

import (
	"context"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func init() {
	// Add our required logging hooks on application initialization.
	//
	// Request ID middleware hook.
	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		if id := RequestID(ctx); id != "" {
			return append(kvs, kv.Field{K: "requestID", V: id})
		}
		return kvs
	})
	// Client IP middleware hook.
	log.Hook(func(ctx context.Context, kvs []kv.Field) []kv.Field {
		if id := PublicKeyID(ctx); id != "" {
			return append(kvs, kv.Field{K: "pubKeyID", V: id})
		}
		return kvs
	})
}
