package gtsmodel

import "time"

// Block refers to the blocking of one account by another.
type Block struct {
	// id of this block in the database
	ID string `pg:"type:CHAR(26),pk,notnull"`
	// When was this block created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this block updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Who created this block?
	AccountID string   `pg:"type:CHAR(26),notnull"`
	Account   *Account `pg:"rel:has-one"`
	// Who is targeted by this block?
	TargetAccountID string   `pg:"type:CHAR(26),notnull"`
	TargetAccount   *Account `pg:"rel:has-one"`
	// Activitypub URI for this block
	URI string `pg:",notnull"`
}
