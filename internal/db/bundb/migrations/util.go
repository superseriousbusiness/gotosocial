// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package migrations

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/dialect/sqltype"
	"github.com/uptrace/bun/schema"
)

// getBunColumnDef generates a column definition string for the SQL table represented by
// Go type, with the SQL column represented by the given Go field name. This ensures when
// adding a new column for table by migration that it will end up as bun would create it.
//
// NOTE: this function must stay in sync with (*bun.CreateTableQuery{}).AppendQuery(),
// specifically where it loops over table fields appending each column definition.
func getBunColumnDef(db bun.IDB, rtype reflect.Type, fieldName string) (string, error) {
	d := db.Dialect()
	f := d.Features()

	// Get bun schema definitions for Go type and its field.
	field, table, err := getModelField(db, rtype, fieldName)
	if err != nil {
		return "", err
	}

	// Start with reasonable buf.
	buf := make([]byte, 0, 64)

	// Start with the SQL column name.
	buf = append(buf, field.SQLName...)
	buf = append(buf, " "...)

	// Append the SQL
	// type information.
	switch {

	// Most of the time these two will match, but for the cases where DiscoveredSQLType is dialect-specific,
	// e.g. pgdialect would change sqltype.SmallInt to pgTypeSmallSerial for columns that have `bun:",autoincrement"`
	case !strings.EqualFold(field.CreateTableSQLType, field.DiscoveredSQLType):
		buf = append(buf, field.CreateTableSQLType...)

	// For all common SQL types except VARCHAR, both UserDefinedSQLType and DiscoveredSQLType specify the correct type,
	// and we needn't modify it. For VARCHAR columns, we will stop to check if a valid length has been set in .Varchar(int).
	case !strings.EqualFold(field.CreateTableSQLType, sqltype.VarChar):
		buf = append(buf, field.CreateTableSQLType...)

	// All else falls back
	// to a default varchar.
	default:
		if d.Name() == dialect.Oracle {
			buf = append(buf, "VARCHAR2"...)
		} else {
			buf = append(buf, sqltype.VarChar...)
		}
		buf = append(buf, sqltype.VarChar+"("...)
		buf = strconv.AppendInt(buf, int64(d.DefaultVarcharLen()), 10)
		buf = append(buf, ")"...)
	}

	// Append not null definition if field requires.
	if field.NotNull && d.Name() != dialect.Oracle {
		buf = append(buf, " NOT NULL"...)
	}

	// Append autoincrement definition if field requires.
	if field.Identity && f.Has(feature.GeneratedIdentity) ||
		(field.AutoIncrement && (f.Has(feature.AutoIncrement) || f.Has(feature.Identity))) {
		buf = d.AppendSequence(buf, table, field)
	}

	// Append any default value.
	if field.SQLDefault != "" {
		buf = append(buf, " DEFAULT "...)
		buf = append(buf, field.SQLDefault...)
	}

	return byteutil.B2S(buf), nil
}

// getModelField returns the uptrace/bun schema details for given Go type and field name.
func getModelField(db bun.IDB, rtype reflect.Type, fieldName string) (*schema.Field, *schema.Table, error) {

	// Get the associated table for Go type.
	table := db.Dialect().Tables().Get(rtype)
	if table == nil {
		return nil, nil, fmt.Errorf("no table found for type: %s", rtype)
	}

	var field *schema.Field

	// Look for field matching Go name.
	for i := range table.Fields {
		if table.Fields[i].GoName == fieldName {
			field = table.Fields[i]
			break
		}
	}

	if field == nil {
		return nil, nil, fmt.Errorf("no bun field found on %s with name: %s", rtype, fieldName)
	}

	return field, table, nil
}

// doesColumnExist safely checks whether given column exists on table, handling both SQLite and PostgreSQL appropriately.
func doesColumnExist(ctx context.Context, tx bun.Tx, table, col string) (bool, error) {
	var n int
	var err error
	switch tx.Dialect().Name() {
	case dialect.SQLite:
		err = tx.NewRaw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?", table, col).Scan(ctx, &n)
	case dialect.PG:
		err = tx.NewRaw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name=? and column_name=?", table, col).Scan(ctx, &n)
	default:
		panic("unexpected dialect")
	}
	return (n > 0), err
}
