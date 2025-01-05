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

package subscriptions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/subscriptions"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

const (
	rMediaPath    = "../../testrig/media"
	rTemplatePath = "../../web/template"
)

type SubscriptionsTestSuite struct {
	suite.Suite

	testAccounts map[string]*gtsmodel.Account
}

func (suite *SubscriptionsTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksCSV() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a CSV list of baddies.
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// We should now have blocks for
	// each domain on the subscribed list.
	for _, domain := range []string{
		"bumfaces.net",
		"peepee.poopoo",
		"nothanks.com",
	} {
		var (
			perm gtsmodel.DomainPermission
			err  error
		)
		if !testrig.WaitFor(func() bool {
			perm, err = testStructs.State.DB.GetDomainBlock(ctx, domain)
			return err == nil
		}) {
			suite.FailNowf("", "timed out waiting for domain %s", domain)
		}

		suite.Equal(testSubscription.ID, perm.GetSubscriptionID())
	}

	// The just-fetched perm sub should
	// have ETag and count etc set now.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have some perms now.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("bigbums6969", permSub.ETag)
	suite.EqualValues(3, count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.WithinDuration(time.Now(), permSub.SuccessfullyFetchedAt, 1*time.Minute)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksJSON() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a JSON list of baddies.
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.json",
			ContentType:        gtsmodel.DomainPermSubContentTypeJSON,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// We should now have blocks for
	// each domain on the subscribed list.
	for _, domain := range []string{
		"bumfaces.net",
		"peepee.poopoo",
		"nothanks.com",
	} {
		var (
			perm gtsmodel.DomainPermission
			err  error
		)
		if !testrig.WaitFor(func() bool {
			perm, err = testStructs.State.DB.GetDomainBlock(ctx, domain)
			return err == nil
		}) {
			suite.FailNowf("", "timed out waiting for domain %s", domain)
		}

		suite.Equal(testSubscription.ID, perm.GetSubscriptionID())
	}

	// The just-fetched perm sub should
	// have ETag and count etc set now.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have some perms now.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("don't modify me daddy", permSub.ETag)
	suite.EqualValues(3, count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.WithinDuration(time.Now(), permSub.SuccessfullyFetchedAt, 1*time.Minute)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksPlain() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a plain list of baddies.
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.txt",
			ContentType:        gtsmodel.DomainPermSubContentTypePlain,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// We should now have blocks for
	// each domain on the subscribed list.
	for _, domain := range []string{
		"bumfaces.net",
		"peepee.poopoo",
		"nothanks.com",
	} {
		var (
			perm gtsmodel.DomainPermission
			err  error
		)
		if !testrig.WaitFor(func() bool {
			perm, err = testStructs.State.DB.GetDomainBlock(ctx, domain)
			return err == nil
		}) {
			suite.FailNowf("", "timed out waiting for domain %s", domain)
		}

		suite.Equal(testSubscription.ID, perm.GetSubscriptionID())
	}

	// The just-fetched perm sub should
	// have ETag and count etc set now.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have some perms now.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("this is a legit etag i swear", permSub.ETag)
	suite.EqualValues(3, count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.WithinDuration(time.Now(), permSub.SuccessfullyFetchedAt, 1*time.Minute)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksCSVETag() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a CSV list of baddies.
		// Include the ETag so it gets sent with the request.
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
			ETag:               "bigbums6969",
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// We should now NOT have blocks for the domains
	// on the list, as the remote will have returned
	// 304, indicating we should do nothing.
	for _, domain := range []string{
		"bumfaces.net",
		"peepee.poopoo",
		"nothanks.com",
	} {
		_, err := testStructs.State.DB.GetDomainBlock(ctx, domain)
		if !errors.Is(err, db.ErrNoEntries) {
			suite.FailNowf("", "domain perm %s created when it shouldn't be")
		}
	}

	// The just-fetched perm sub should
	// have ETag and count etc set now.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have no perms.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal("bigbums6969", permSub.ETag)
	suite.Zero(count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.WithinDuration(time.Now(), permSub.SuccessfullyFetchedAt, 1*time.Minute)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocks404() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a CSV list of baddies.
		// The endpoint will return a 404 so we can test erroring.
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/does_not_exist.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// The just-fetched perm sub should have an error set on it.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have no perms.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Zero(count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.Zero(permSub.SuccessfullyFetchedAt)
	suite.Equal(`DereferenceDomainPermissions: GET request to https://lists.example.org/does_not_exist.csv failed: status="" body="{"error":"not found"}"`, permSub.Error)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksWrongContentTypeCSV() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a plaintext list of baddies,
		// but try to parse as CSV content type (shouldn't work).
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.txt",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// The just-fetched perm sub should have an error set on it.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have no perms.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Zero(count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.Zero(permSub.SuccessfullyFetchedAt)
	suite.Equal(`permsFromCSV: unexpected column headers in csv: [bumfaces.net]`, permSub.Error)
}

func (suite *SubscriptionsTestSuite) TestDomainBlocksWrongContentTypePlain() {
	var (
		ctx           = context.Background()
		testStructs   = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		testAccount   = suite.testAccounts["admin_account"]
		subscriptions = subscriptions.New(
			testStructs.State,
			testStructs.TransportController,
			testStructs.TypeConverter,
		)

		// Create a subscription for a plaintext list of baddies,
		// but try to parse as CSV content type (shouldn't work).
		testSubscription = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGE681TQSBPAV59GZXPKE62H",
			Priority:           255,
			Title:              "whatever!",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			AsDraft:            util.Ptr(false),
			AdoptOrphans:       util.Ptr(true),
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://lists.example.org/baddies.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypePlain,
		}
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Store test subscription.
	if err := testStructs.State.DB.PutDomainPermissionSubscription(
		ctx, testSubscription,
	); err != nil {
		suite.FailNow(err.Error())
	}

	// Process all subscriptions.
	subscriptions.ProcessDomainPermissionSubscriptions(ctx, testSubscription.PermissionType)

	// The just-fetched perm sub should have an error set on it.
	permSub, err := testStructs.State.DB.GetDomainPermissionSubscriptionByID(
		ctx, testSubscription.ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Should have no perms.
	count, err := testStructs.State.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Zero(count)
	suite.WithinDuration(time.Now(), permSub.FetchedAt, 1*time.Minute)
	suite.Zero(permSub.SuccessfullyFetchedAt)
	suite.Equal(`fetch successful but parsed zero usable results`, permSub.Error)
}

func TestSubscriptionTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionsTestSuite))
}
