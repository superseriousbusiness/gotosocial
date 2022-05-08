package runners

import (
	"context"
	"time"
)

// closedctx is an always closed context.
var closedctx = func() context.Context {
	ctx := make(cancelctx)
	close(ctx)
	return ctx
}()

// ContextWithCancel returns a new context.Context impl with cancel.
func ContextWithCancel() (context.Context, context.CancelFunc) {
	ctx := make(cancelctx)
	return ctx, func() { close(ctx) }
}

// cancelctx is the simplest possible cancellable context.
type cancelctx (chan struct{})

func (cancelctx) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (ctx cancelctx) Done() <-chan struct{} {
	return ctx
}

func (ctx cancelctx) Err() error {
	select {
	case <-ctx:
		return context.Canceled
	default:
		return nil
	}
}

func (cancelctx) Value(key interface{}) interface{} {
	return nil
}
