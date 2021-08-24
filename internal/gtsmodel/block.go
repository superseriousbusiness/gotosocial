package gtsmodel

import "time"

// Block refers to the blocking of one account by another.
type Block struct {
	// id of this block in the database
	ID string `bun:"type:CHAR(26),pk,notnull"`
	// When was this block created
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// When was this block updated
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	// Who created this block?
	AccountID string   `bun:"type:CHAR(26),notnull"`
	Account   *Account `bun:"rel:belongs-to"`
	// Who is targeted by this block?
	TargetAccountID string   `bun:"type:CHAR(26),notnull"`
	TargetAccount   *Account `bun:"rel:belongs-to"`
	// Activitypub URI for this block
	URI string `bun:",notnull"`
}
