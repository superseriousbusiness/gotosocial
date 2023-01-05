# Migrations

## How do I write a migration file?

[See here](https://bun.uptrace.dev/guide/migrations.html#migration-names)

As a template, take one of the existing migration files and modify it, or use the below code snippet:

```go
/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package migrations

import (
    "context"

    "github.com/uptrace/bun"
)

func init() {
    up := func(ctx context.Context, db *bun.DB) error {
        return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
            // your logic here
            return nil
        })
    }

    down := func(ctx context.Context, db *bun.DB) error {
        return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
            // your logic here
            return nil
        })
    }

    if err := Migrations.Register(up, down); err != nil {
        panic(err)
    }
}
```

## File format

Bun requires a very specific format: 14 digits, then letters or underscores.

You can use the following bash command on your branch to generate a suitable migration filename.

```bash
echo "$(date --utc +%Y%m%d%H%M%S | head -c 14)_$(git rev-parse --abbrev-ref HEAD).go"
```

## Rules of thumb

1. **DON'T DROP TABLES**!!!!!!!!
2. Don't make something `NOT NULL` if it's likely to already contain `null` fields.
