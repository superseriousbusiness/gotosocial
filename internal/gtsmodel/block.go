package gtsmodel

import "time"

// Block refers to the blocking of one account by another.
type Block struct {
	ID              string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`    // id of this item in the database
	CreatedAt       time.Time `validate:"-" bun:",nullzero,notnull,default:current_timestamp"`             // when was item created
	UpdatedAt       time.Time `validate:"-" bun:",nullzero,notnull,default:current_timestamp"`             // when was item last updated
	URI             string    `validate:"required,url" bun:",notnull,nullzero,unique"`                     // ActivityPub uri of this block.
	AccountID       string    `validate:"required,ulid" bun:"type:CHAR(26),unique:blocksrctarget,notnull"` // Who does this block originate from?
	Account         *Account  `validate:"-" bun:"rel:belongs-to"`                                          // Account corresponding to accountID
	TargetAccountID string    `validate:"required,ulid" bun:"type:CHAR(26),unique:blocksrctarget,notnull"` // Who is the target of this block ?
	TargetAccount   *Account  `validate:"-" bun:"rel:belongs-to"`                                          // Account corresponding to targetAccountID
}
