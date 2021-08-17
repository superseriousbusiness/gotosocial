package db

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Relationship interface {
	// Blocked checks whether a block exists in eiher direction between two accounts.
	// That is, it returns true if account1 blocks account2, OR if account2 blocks account1.
	Blocked(account1 string, account2 string) (bool, DBError)

	// GetRelationship retrieves the relationship of the targetAccount to the requestingAccount.
	GetRelationship(requestingAccount string, targetAccount string) (*gtsmodel.Relationship, DBError)

	// Follows returns true if sourceAccount follows target account, or an error if something goes wrong while finding out.
	Follows(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, DBError)

	// FollowRequested returns true if sourceAccount has requested to follow target account, or an error if something goes wrong while finding out.
	FollowRequested(sourceAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (bool, DBError)

	// Mutuals returns true if account1 and account2 both follow each other, or an error if something goes wrong while finding out.
	Mutuals(account1 *gtsmodel.Account, account2 *gtsmodel.Account) (bool, DBError)

	// AcceptFollowRequest moves a follow request in the database from the follow_requests table to the follows table.
	// In other words, it should create the follow, and delete the existing follow request.
	//
	// It will return the newly created follow for further processing.
	AcceptFollowRequest(originAccountID string, targetAccountID string) (*gtsmodel.Follow, DBError)
}
