package gtscontext

import "context"

// package private context key type.
type ctxkey uint

const (
	// context keys.
	_ ctxkey = iota
	barebonesKey
)

// Barebones returns whether the "barebones" context key has been set. This
// can be used to indicate to the database, for example, that only a barebones
// model need be returned, Allowing it to skip populating sub models.
func Barebones(ctx context.Context) bool {
	_, ok := ctx.Value(barebonesKey).(struct{})
	return ok
}

// SetBarebones sets the "barebones" context flag and returns this wrapped context.
// See Barebones() for further information on the "barebones" context flag..
func SetBarebones(ctx context.Context) context.Context {
	return context.WithValue(ctx, barebonesKey, struct{}{})
}
