package otelsql

import (
	"context"
	"database/sql/driver"

	"go.opentelemetry.io/otel/trace"
)

type otelTx struct {
	ctx     context.Context
	tx      driver.Tx
	instrum *dbInstrum
}

var _ driver.Tx = (*otelTx)(nil)

func newTx(ctx context.Context, tx driver.Tx, instrum *dbInstrum) *otelTx {
	return &otelTx{
		ctx:     ctx,
		tx:      tx,
		instrum: instrum,
	}
}

func (tx *otelTx) Commit() error {
	return tx.instrum.withSpan(tx.ctx, "tx.Commit", "",
		func(ctx context.Context, span trace.Span) error {
			return tx.tx.Commit()
		})
}

func (tx *otelTx) Rollback() error {
	return tx.instrum.withSpan(tx.ctx, "tx.Rollback", "",
		func(ctx context.Context, span trace.Span) error {
			return tx.tx.Rollback()
		})
}
