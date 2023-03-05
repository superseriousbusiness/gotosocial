package gtscontext

import "context"

// package private context key type.
type ctxkey uint

const (
	// context keys.
	_ ctxkey = iota
	barebonesKey
	calledKey
)

func Barebones(ctx context.Context) bool {
	_, ok := ctx.Value(barebonesKey).(struct{})
	return ok
}

func SetBarebones(ctx context.Context) context.Context {
	return context.WithValue(ctx, barebonesKey, struct{}{})
}

func Called(ctx context.Context) bool {
	_, ok := ctx.Value(calledKey).(struct{})
	return ok
}

func SetCalled(ctx context.Context) context.Context {
	return context.WithValue(ctx, calledKey, struct{}{})
}
