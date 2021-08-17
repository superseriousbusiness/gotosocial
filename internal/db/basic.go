package db

import "context"

type Basic interface {
	// CreateTable creates a table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	CreateTable(i interface{}) DBError

	// DropTable drops the table for the given interface.
	// For implementations that don't use tables, this can just return nil.
	DropTable(i interface{}) DBError

	// Stop should stop and close the database connection cleanly, returning an error if this is not possible.
	// If the database implementation doesn't need to be stopped, this can just return nil.
	Stop(ctx context.Context) DBError

	// IsHealthy should return nil if the database connection is healthy, or an error if not.
	IsHealthy(ctx context.Context) DBError

	// GetByID gets one entry by its id. In a database like postgres, this might be the 'id' field of the entry,
	// for other implementations (for example, in-memory) it might just be the key of a map.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetByID(id string, i interface{}) DBError

	// GetWhere gets one entry where key = value. This is similar to GetByID but allows the caller to specify the
	// name of the key to select from.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetWhere(where []Where, i interface{}) DBError

	// GetAll will try to get all entries of type i.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	// In case of no entries, a 'no entries' error will be returned
	GetAll(i interface{}) DBError

	// Put simply stores i. It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Put(i interface{}) DBError

	// Upsert stores or updates i based on the given conflict column, as in https://www.postgresqltutorial.com/postgresql-upsert/
	// It is up to the implementation to figure out how to store it, and using what key.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	Upsert(i interface{}, conflictColumn string) DBError

	// UpdateByID updates i with id id.
	// The given interface i will be set to the result of the query, whatever it is. Use a pointer or a slice.
	UpdateByID(id string, i interface{}) DBError

	// UpdateOneByID updates interface i with database the given database id. It will update one field of key key and value value.
	UpdateOneByID(id string, key string, value interface{}, i interface{}) DBError

	// UpdateWhere updates column key of interface i with the given value, where the given parameters apply.
	UpdateWhere(where []Where, key string, value interface{}, i interface{}) DBError

	// DeleteByID removes i with id id.
	// If i didn't exist anyway, then no error should be returned.
	DeleteByID(id string, i interface{}) DBError

	// DeleteWhere deletes i where key = value
	// If i didn't exist anyway, then no error should be returned.
	DeleteWhere(where []Where, i interface{}) DBError
}
