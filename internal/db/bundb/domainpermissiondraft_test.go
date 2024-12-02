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

package bundb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type DomainPermissionDraftTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DomainPermissionDraftTestSuite) TestPermDraftCreateGetDelete() {
	var (
		ctx   = context.Background()
		draft = &gtsmodel.DomainPermissionDraft{
			ID:                 "01JCZN614XG85GCGAMSV9ZZAEJ",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			Domain:             "exämple.org",
			CreatedByAccountID: suite.testAccounts["admin_account"].ID,
			PrivateComment:     "this domain is poo",
			PublicComment:      "this domain is poo, but phrased in a more outward-facing way",
			Obfuscate:          util.Ptr(false),
			SubscriptionID:     "01JCZN8PG55KKEVTDAY52D0T3P",
		}
	)

	// Whack the draft in.
	if err := suite.state.DB.PutDomainPermissionDraft(ctx, draft); err != nil {
		suite.FailNow(err.Error())
	}

	// Get the draft again.
	dbDraft, err := suite.state.DB.GetDomainPermissionDraftByID(ctx, draft.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Domain should have been stored punycoded.
	suite.Equal("xn--exmple-cua.org", dbDraft.Domain)

	// Search for domain using both
	// punycode and unicode variants.
	search1, err := suite.state.DB.GetDomainPermissionDrafts(
		ctx,
		gtsmodel.DomainPermissionUnknown,
		"",
		"exämple.org",
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if len(search1) != 1 {
		suite.FailNow("couldn't get domain perm draft exämple.org")
	}

	search2, err := suite.state.DB.GetDomainPermissionDrafts(
		ctx,
		gtsmodel.DomainPermissionUnknown,
		"",
		"xn--exmple-cua.org",
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if len(search2) != 1 {
		suite.FailNow("couldn't get domain perm draft example.org")
	}

	// Change ID + try to put the same draft again.
	draft.ID = "01JCZNVYSDT3JE385FABMJ7ADQ"
	err = suite.state.DB.PutDomainPermissionDraft(ctx, draft)
	if !errors.Is(err, db.ErrAlreadyExists) {
		suite.FailNow("was able to insert same domain perm draft twice")
	}

	// Put same draft but change permission type, should work.
	draft.PermissionType = gtsmodel.DomainPermissionAllow
	if err := suite.state.DB.PutDomainPermissionDraft(ctx, draft); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete both drafts.
	for _, id := range []string{
		"01JCZN614XG85GCGAMSV9ZZAEJ",
		"01JCZNVYSDT3JE385FABMJ7ADQ",
	} {
		if err := suite.state.DB.DeleteDomainPermissionDraft(ctx, id); err != nil {
			suite.FailNow("error deleting domain permission draft")
		}
	}
}

func TestDomainPermissionDraftTestSuite(t *testing.T) {
	suite.Run(t, new(DomainPermissionDraftTestSuite))
}
