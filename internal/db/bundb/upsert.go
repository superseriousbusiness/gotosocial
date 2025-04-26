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

package bundb

import (
	"context"
	"database/sql"
	"reflect"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

// UpsertQuery is a wrapper around an insert query that can update if an insert fails.
// Doesn't implement the full set of Bun query methods, but we can add more if we need them.
// See https://bun.uptrace.dev/guide/query-insert.html#upsert
type UpsertQuery struct {
	db          bun.IDB
	model       interface{}
	constraints []string
	columns     []string
}

func NewUpsert(idb bun.IDB) *UpsertQuery {
	// note: passing in rawtx as conn iface so no double query-hook
	// firing when passed through the bun.Tx.Query___() functions.
	return &UpsertQuery{db: idb}
}

// Model sets the model or models to upsert.
func (u *UpsertQuery) Model(model interface{}) *UpsertQuery {
	u.model = model
	return u
}

// Constraint sets the columns or indexes that are used to check for conflicts.
// This is required.
func (u *UpsertQuery) Constraint(constraints ...string) *UpsertQuery {
	u.constraints = constraints
	return u
}

// Column sets the columns to update if an insert does't happen.
// If empty, all columns not being used for constraints will be updated.
// Cannot overlap with Constraint.
func (u *UpsertQuery) Column(columns ...string) *UpsertQuery {
	u.columns = columns
	return u
}

// insertDialect errors if we're using a dialect in which we don't know how to upsert.
func (u *UpsertQuery) insertDialect() error {
	dialectName := u.db.Dialect().Name()
	switch dialectName {
	case dialect.PG, dialect.SQLite:
		return nil
	default:
		// FUTURE: MySQL has its own variation on upserts, but the syntax is different.
		return gtserror.Newf("UpsertQuery: upsert not supported by SQL dialect: %s", dialectName)
	}
}

// insertConstraints checks that we have constraints and returns them.
func (u *UpsertQuery) insertConstraints() ([]string, error) {
	if len(u.constraints) == 0 {
		return nil, gtserror.New("UpsertQuery: upserts require at least one constraint column or index, none provided")
	}
	return u.constraints, nil
}

// insertColumns returns the non-constraint columns we'll be updating.
func (u *UpsertQuery) insertColumns(constraints []string) ([]string, error) {
	// Constraints as a set.
	constraintSet := make(map[string]struct{}, len(constraints))
	for _, constraint := range constraints {
		constraintSet[constraint] = struct{}{}
	}

	var columns []string
	var err error
	if len(u.columns) == 0 {
		columns, err = u.insertColumnsDefault(constraintSet)
	} else {
		columns, err = u.insertColumnsSpecified(constraintSet)
	}
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, gtserror.New("UpsertQuery: there are no columns to update when upserting")
	}

	return columns, nil
}

// hasElem returns whether the type has an element and can call [reflect.Type.Elem] without panicking.
func hasElem(modelType reflect.Type) bool {
	switch modelType.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	default:
		return false
	}
}

// insertColumnsDefault returns all non-constraint columns from the model schema.
func (u *UpsertQuery) insertColumnsDefault(constraintSet map[string]struct{}) ([]string, error) {
	// Get underlying struct type.
	modelType := reflect.TypeOf(u.model)
	for hasElem(modelType) {
		modelType = modelType.Elem()
	}

	table := u.db.Dialect().Tables().Get(modelType)
	if table == nil {
		return nil, gtserror.Newf("UpsertQuery: couldn't find the table schema for model: %v", u.model)
	}

	columns := make([]string, 0, len(u.columns))
	for _, field := range table.Fields {
		column := field.Name
		if _, overlaps := constraintSet[column]; !overlaps {
			columns = append(columns, column)
		}
	}

	return columns, nil
}

// insertColumnsSpecified ensures constraints and specified columns to update don't overlap.
func (u *UpsertQuery) insertColumnsSpecified(constraintSet map[string]struct{}) ([]string, error) {
	overlapping := make([]string, 0, min(len(u.constraints), len(u.columns)))
	for _, column := range u.columns {
		if _, overlaps := constraintSet[column]; overlaps {
			overlapping = append(overlapping, column)
		}
	}

	if len(overlapping) > 0 {
		return nil, gtserror.Newf(
			"UpsertQuery: the following columns can't be used for both constraints and columns to update: %s",
			strings.Join(overlapping, ", "),
		)
	}

	return u.columns, nil
}

// insert tries to create a Bun insert query from an upsert query.
func (u *UpsertQuery) insertQuery() (*bun.InsertQuery, error) {
	var err error

	err = u.insertDialect()
	if err != nil {
		return nil, err
	}

	constraints, err := u.insertConstraints()
	if err != nil {
		return nil, err
	}

	columns, err := u.insertColumns(constraints)
	if err != nil {
		return nil, err
	}

	// Build the parts of the query that need us to generate SQL.
	constraintIDPlaceholders := make([]string, 0, len(constraints))
	constraintIDs := make([]interface{}, 0, len(constraints))
	for _, constraint := range constraints {
		constraintIDPlaceholders = append(constraintIDPlaceholders, "?")
		constraintIDs = append(constraintIDs, bun.Ident(constraint))
	}
	onSQL := "CONFLICT (" + strings.Join(constraintIDPlaceholders, ", ") + ") DO UPDATE"

	setClauses := make([]string, 0, len(columns))
	setIDs := make([]interface{}, 0, 2*len(columns))
	for _, column := range columns {
		setClauses = append(setClauses, "? = ?")
		// "excluded" is a special table that contains only the row involved in a conflict.
		setIDs = append(setIDs, bun.Ident(column), bun.Ident("excluded."+column))
	}
	setSQL := strings.Join(setClauses, ", ")

	insertQuery := u.db.
		NewInsert().
		Model(u.model).
		On(onSQL, constraintIDs...).
		Set(setSQL, setIDs...)

	return insertQuery, nil
}

// Exec builds a Bun insert query from the upsert query, and executes it.
func (u *UpsertQuery) Exec(ctx context.Context, dest ...interface{}) (sql.Result, error) {
	insertQuery, err := u.insertQuery()
	if err != nil {
		return nil, err
	}

	return insertQuery.Exec(ctx, dest...)
}

// Scan builds a Bun insert query from the upsert query, and scans it.
func (u *UpsertQuery) Scan(ctx context.Context, dest ...interface{}) error {
	insertQuery, err := u.insertQuery()
	if err != nil {
		return err
	}

	return insertQuery.Scan(ctx, dest...)
}
