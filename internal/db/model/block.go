package model

import "time"

// Block refers to the blocking of one account by another.
type Block struct {
	// id of this block in the database
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	// When was this block created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// When was this block updated
	UpdatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// Who created this block?
	AccountID string `pg:",notnull"`
	// Who is targeted by this block?
	TargetAccountID string `pg:",notnull"`
	// Activitypub URI for this block
	URI string
}
