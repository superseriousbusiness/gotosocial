/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

import "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

type Relationship interface {
	// Blocked checks whether account 1 has a block in place against block2.
	// If eitherDirection is true, then the function returns true if account1 blocks account2, OR if account2 blocks account1.
	Blocked(account1 string, account2 string, eitherDirection bool) (bool, DBError)

	// GetBlock returns the block from account1 targeting account2, if it exists, or an error if it doesn't.
	//
	// Because this is slower than Blocked, only use it if you need the actual Block struct for some reason,
	// not if you're just checking for the existence of a block.
	GetBlock(account1 string, account2 string) (*gtsmodel.Block, DBError)

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
