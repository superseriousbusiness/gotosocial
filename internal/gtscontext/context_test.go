package gtscontext_test

import (
	"context"
	"net/url"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

func BenchmarkContexts(b *testing.B) {
	var receiving *gtsmodel.Account
	var requesting *gtsmodel.Account
	var otherIRIs []*url.URL

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()

			ctx = gtscontext.SetBarebones(ctx)
			ctx = gtscontext.SetFastFail(ctx)
			ctx = gtscontext.SetDryRun(ctx)
			ctx = gtscontext.SetReceivingAccount(ctx, receiving)
			ctx = gtscontext.SetRequestingAccount(ctx, requesting)
			ctx = gtscontext.SetOtherIRIs(ctx, otherIRIs)

			if !gtscontext.Barebones(ctx) {
				println("oh no!")
			}

			if !gtscontext.IsFastfail(ctx) {
				println("oh no!")
			}

			if !gtscontext.DryRun(ctx) {
				println("oh no!")
			}

			if gtscontext.ReceivingAccount(ctx) != nil {
				println("oh no!")
			}

			if gtscontext.RequestingAccount(ctx) != nil {
				println("oh no!")
			}

			if len(gtscontext.OtherIRIs(ctx)) > 0 {
				println("oh no!")
			}
		}
	})
}
