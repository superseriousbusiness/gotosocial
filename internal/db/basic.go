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

package db

import "context"

// Basic wraps basic database functionality.
type Basic interface {
	// CreateTable creates a table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	CreateTable(ctx context.Context, i interface{}) Error

	// CreateAllTables creates *all* tables necessary for the running of GoToSocial.
	// Because it uses the 'if not exists' parameter it is safe to run against a GtS that's already been initialized.
	CreateAllTables(ctx context.Context) Error

	// DropTable drops the table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	DropTable(ctx context.Context, i interface{}) Error

	// Stop should stop and close the database connection cleanly, returning an error if this is not possible.
	// If the database implementation doesn't need to be stopped, this can just return nil.
	Stop(ctx context.Context) Error

	// IsHealthy should return nil if the database connection is healthy, or an error if not.
	IsHealthy(ctx context.Context) Error

	// GetByID gets one entry by its id. In a database like postgres, this might be the 'id' field of the entry,
	// for other implementations (for example, in-memory) it might just be the key of a map.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetByID(ctx context.Context, id string, i interface{}) Error

	// GetWhere gets one entry where key = value. This is similar to GetByID but allows the caller to specify the
	// name of the key to select from.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetWhere(ctx context.Context, where []Where, i interface{}) Error

	// GetAll will try to get all entries of type i.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetAll(ctx context.Context, i interface{}) Error

	// Put simply stores i. It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Put(ctx context.Context, i interface{}) Error

	// UpdateByID updates values of i based on its id.
	// If any columns are specified, these will be updated exclusively.
	// Otherwise, the whole model will be updated.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	UpdateByID(ctx context.Context, i interface{}, id string, columns ...string) Error

	// UpdateWhere updates column key of interface i with the given value, where the given parameters apply.
	UpdateWhere(ctx context.Context, where []Where, key string, value interface{}, i interface{}) Error

	// DeleteByID removes i with id id.
	// If i didn't exist anyway, then no error should be returned.
	DeleteByID(ctx context.Context, id string, i interface{}) Error

	// DeleteWhere deletes i where key = value
	// If i didn't exist anyway, then no error should be returned.
	DeleteWhere(ctx context.Context, where []Where, i interface{}) Error
}
