package gtsmodel

import "time"

// Instance represents a federated instance, either local or remote.
type Instance struct {
	// ID of this instance in the database
	ID string `pg:"type:CHAR(26),pk,notnull,unique"`
	// Instance domain eg example.org
	Domain string `pg:",pk,notnull,unique"`
	// Title of this instance as it would like to be displayed.
	Title string
	// base URI of this instance eg https://example.org
	URI string `pg:",notnull,unique"`
	// When was this instance created in the db?
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this instance last updated in the db?
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this instance suspended, if at all?
	SuspendedAt time.Time
	// ID of any existing domain block for this instance in the database
	DomainBlockID string       `pg:"type:CHAR(26)"`
	DomainBlock   *DomainBlock `pg:"rel:has-one"`
	// Short description of this instance
	ShortDescription string
	// Longer description of this instance
	Description string
	// Terms and conditions of this instance
	Terms string
	// Contact email address for this instance
	ContactEmail string
	// Username of the contact account for this instance
	ContactAccountUsername string
	// Contact account ID in the database for this instance
	ContactAccountID string   `pg:"type:CHAR(26)"`
	ContactAccount   *Account `pg:"rel:has-one"`
	// Reputation score of this instance
	Reputation int64 `pg:",notnull,default:0"`
	// Version of the software used on this instance
	Version string
}
