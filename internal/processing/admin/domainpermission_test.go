// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package admin_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type DomainBlockTestSuite struct {
	AdminStandardTestSuite
}

type domainPermAction struct {
	// 'create' or 'delete'
	// the domain permission.
	createOrDelete string

	// Type of permission
	// to create or delete.
	permissionType gtsmodel.DomainPermissionType

	// Domain to target
	// with the permission.
	domain string

	// Expected result of this
	// permission action on each
	// account on the target domain.
	// Eg., suite.Zero(account.SuspendedAt)
	expected func(*gtsmodel.Account) bool
}

type domainPermTest struct {
	// Federation mode under which to
	// run this test. This is important
	// because it may effect which side
	// effects are taken, if any.
	instanceFederationMode string

	// Series of actions to run as part
	// of this test. After each action,
	// expected will be called. This
	// allows testers to run multiple
	// actions in a row and check that
	// the results after each action are
	// what they expected, in light of
	// previous actions.
	actions []domainPermAction
}

// run a domainPermTest by running each of
// its actions in turn and checking results.
func (suite *DomainBlockTestSuite) runDomainPermTest(t domainPermTest) {
	config.SetInstanceFederationMode(t.instanceFederationMode)

	for _, action := range t.actions {
		// Run the desired action.
		var actionID string
		switch action.createOrDelete {
		case "create":
			_, actionID = suite.createDomainPerm(action.permissionType, action.domain)
		case "delete":
			_, actionID = suite.deleteDomainPerm(action.permissionType, action.domain)
		default:
			panic("createOrDelete was not 'create' or 'delete'")
		}

		// Let the action finish.
		suite.awaitAction(actionID)

		// Check expected results
		// against each account.
		accounts, err := suite.db.GetInstanceAccounts(
			context.Background(),
			action.domain,
			"", 0,
		)
		if err != nil {
			suite.FailNow("", "error getting instance accounts for %s: %v", action.domain, err)
		}

		for _, account := range accounts {
			if !action.expected(account) {
				suite.T().FailNow()
			}
		}
	}
}

// create given permissionType with default values.
func (suite *DomainBlockTestSuite) createDomainPerm(
	permissionType gtsmodel.DomainPermissionType,
	domain string,
) (*apimodel.DomainPermission, string) {
	ctx := context.Background()

	apiPerm, actionID, errWithCode := suite.adminProcessor.DomainPermissionCreate(
		ctx,
		permissionType,
		suite.testAccounts["admin_account"],
		domain,
		false,
		"",
		"",
		"",
	)
	suite.NoError(errWithCode)
	suite.NotNil(apiPerm)
	suite.NotEmpty(actionID)

	return apiPerm, actionID
}

// delete given permission type.
func (suite *DomainBlockTestSuite) deleteDomainPerm(
	permissionType gtsmodel.DomainPermissionType,
	domain string,
) (*apimodel.DomainPermission, string) {
	var (
		ctx              = context.Background()
		domainPermission gtsmodel.DomainPermission
	)

	// To delete the permission,
	// first get it from the db.
	switch permissionType {
	case gtsmodel.DomainPermissionBlock:
		domainPermission, _ = suite.db.GetDomainBlock(ctx, domain)
	case gtsmodel.DomainPermissionAllow:
		domainPermission, _ = suite.db.GetDomainAllow(ctx, domain)
	default:
		panic("unrecognized permission type")
	}

	if domainPermission == nil {
		suite.FailNow("domain permission was nil")
	}

	// Now use the ID to delete it.
	apiPerm, actionID, errWithCode := suite.adminProcessor.DomainPermissionDelete(
		ctx,
		permissionType,
		suite.testAccounts["admin_account"],
		domainPermission.GetID(),
	)
	suite.NoError(errWithCode)
	suite.NotNil(apiPerm)
	suite.NotEmpty(actionID)

	return apiPerm, actionID
}

// waits for given actionID to be completed.
func (suite *DomainBlockTestSuite) awaitAction(actionID string) {
	ctx := context.Background()

	if !testrig.WaitFor(func() bool {
		return suite.adminProcessor.Actions().TotalRunning() == 0
	}) {
		suite.FailNow("timed out waiting for admin action(s) to finish")
	}

	// Ensure action marked as
	// completed in the database.
	adminAction, err := suite.db.GetAdminAction(ctx, actionID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotZero(adminAction.CompletedAt)
	suite.Empty(adminAction.Errors)
}

func (suite *DomainBlockTestSuite) TestBlockAndUnblockDomain() {
	const domain = "fossbros-anonymous.io"

	suite.runDomainPermTest(domainPermTest{
		instanceFederationMode: config.InstanceFederationModeBlocklist,
		actions: []domainPermAction{
			{
				createOrDelete: "create",
				permissionType: gtsmodel.DomainPermissionBlock,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Domain was blocked, so each
					// account should now be suspended.
					return suite.NotZero(account.SuspendedAt)
				},
			},
			{
				createOrDelete: "delete",
				permissionType: gtsmodel.DomainPermissionBlock,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Domain was unblocked, so each
					// account should now be unsuspended.
					return suite.Zero(account.SuspendedAt)
				},
			},
		},
	})
}

func (suite *DomainBlockTestSuite) TestBlockAndAllowDomain() {
	const domain = "fossbros-anonymous.io"

	suite.runDomainPermTest(domainPermTest{
		instanceFederationMode: config.InstanceFederationModeBlocklist,
		actions: []domainPermAction{
			{
				createOrDelete: "create",
				permissionType: gtsmodel.DomainPermissionBlock,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Domain was blocked, so each
					// account should now be suspended.
					return suite.NotZero(account.SuspendedAt)
				},
			},
			{
				createOrDelete: "create",
				permissionType: gtsmodel.DomainPermissionAllow,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Domain was explicitly allowed, so each
					// account should now be unsuspended, since
					// the allow supercedes the block.
					return suite.Zero(account.SuspendedAt)
				},
			},
			{
				createOrDelete: "delete",
				permissionType: gtsmodel.DomainPermissionAllow,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Deleting the allow now, while there's
					// still a block in place, should cause
					// the block to take effect again.
					return suite.NotZero(account.SuspendedAt)
				},
			},
			{
				createOrDelete: "delete",
				permissionType: gtsmodel.DomainPermissionBlock,
				domain:         domain,
				expected: func(account *gtsmodel.Account) bool {
					// Deleting the block now should
					// unsuspend the accounts again.
					return suite.Zero(account.SuspendedAt)
				},
			},
		},
	})
}

func TestDomainBlockTestSuite(t *testing.T) {
	suite.Run(t, new(DomainBlockTestSuite))
}
