package migrations

import (
	"context"

	gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Note table.
			if _, err := tx.
				NewCreateTable().
				Model(&gtsmodel.Note{}).
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			// Add IDs index to the Note table.
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Note{}).
				Index("notes_account_id_target_account_id_idx").
				Column("account_id", "target_account_id").
				Exec(ctx); err != nil {
				return err
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return nil
		})
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
