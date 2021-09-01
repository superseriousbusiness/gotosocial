package gtsmodel

import "time"

// Instance represents a federated instance, either local or remote.
type Instance struct {
	ID                     string       `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                     // id of this item in the database
	CreatedAt              time.Time    `validate:"-" bun:",nullzero,notnull,default:current_timestamp"`                              // when was item created
	UpdatedAt              time.Time    `validate:"-" bun:",nullzero,notnull,default:current_timestamp"`                              // when was item last updated
	Domain                 string       `validate:"required,fqdn" bun:",nullzero,notnull,unique"`                                     // Instance domain eg example.org
	Title                  string       `validate:"-" bun:",nullzero"`                                                                // Title of this instance as it would like to be displayed.
	URI                    string       `validate:"required,url" bun:",nullzero,notnull,unique"`                                      // base URI of this instance eg https://example.org
	SuspendedAt            time.Time    `validate:"-" bun:",nullzero"`                                                                // When was this instance suspended, if at all?
	DomainBlockID          string       `validate:"omitempty,ulid" bun:"type:CHAR(26),nullzero"`                                      // ID of any existing domain block for this instance in the database
	DomainBlock            *DomainBlock `validate:"-" bun:"rel:belongs-to"`                                                           // Domain block corresponding to domainBlockID
	ShortDescription       string       `validate:"-" bun:",nullzero"`                                                                // Short description of this instance
	Description            string       `validate:"-" bun:",nullzero"`                                                                // Longer description of this instance
	Terms                  string       `validate:"-" bun:",nullzero"`                                                                // Terms and conditions of this instance
	ContactEmail           string       `validate:"omitempty,email" bun:",nullzero"`                                                  // Contact email address for this instance
	ContactAccountUsername string       `validate:"required_with=ContactAccountID" bun:",nullzero"`                                   // Username of the contact account for this instance
	ContactAccountID       string       `validate:"required_with=ContactAccountUsername,omitempty,ulid" bun:"type:CHAR(26),nullzero"` // Contact account ID in the database for this instance
	ContactAccount         *Account     `validate:"-" bun:"rel:belongs-to"`                                                           // account corresponding to contactAccountID
	Reputation             int64        `validate:"-" bun:",notnull,default:0"`                                                       // Reputation score of this instance
	Version                string       `validate:"-" bun:",nullzero"`                                                                // Version of the software used on this instance
}
