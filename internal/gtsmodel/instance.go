package gtsmodel

import "time"

// Instance represents a federated instance, either local or remote.
type Instance struct {
	// ID of this instance in the database
	ID string `bun:"type:CHAR(26),pk,notnull,unique"`
	// Instance domain eg example.org
	Domain string `bun:",pk,notnull,unique"`
	// Title of this instance as it would like to be displayed.
	Title string `bun:",nullzero"`
	// base URI of this instance eg https://example.org
	URI string `bun:",notnull,unique"`
	// When was this instance created in the db?
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this instance last updated in the db?
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this instance suspended, if at all?
	SuspendedAt time.Time `bun:",nullzero"`
	// ID of any existing domain block for this instance in the database
	DomainBlockID string       `bun:"type:CHAR(26),nullzero"`
	DomainBlock   *DomainBlock `bun:"rel:belongs-to"`
	// Short description of this instance
	ShortDescription string `bun:",nullzero"`
	// Longer description of this instance
	Description string `bun:",nullzero"`
	// Terms and conditions of this instance
	Terms string `bun:",nullzero"`
	// Contact email address for this instance
	ContactEmail string `bun:",nullzero"`
	// Username of the contact account for this instance
	ContactAccountUsername string `bun:",nullzero"`
	// Contact account ID in the database for this instance
	ContactAccountID string   `bun:"type:CHAR(26),nullzero"`
	ContactAccount   *Account `bun:"rel:belongs-to"`
	// Reputation score of this instance
	Reputation int64 `bun:",nullzero,notnull,default:0"`
	// Version of the software used on this instance
	Version string `bun:",nullzero"`
}
