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
)

type DomainPermissionExcludeTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DomainPermissionExcludeTestSuite) TestPermExcludeCreateGetDelete() {
	var (
		ctx     = context.Background()
		exclude = &gtsmodel.DomainPermissionExclude{
			ID:                 "01JCZN614XG85GCGAMSV9ZZAEJ",
			Domain:             "exämple.org",
			CreatedByAccountID: suite.testAccounts["admin_account"].ID,
			PrivateComment:     "this domain is poo",
		}
	)

	// Whack the exclude in.
	if err := suite.state.DB.PutDomainPermissionExclude(ctx, exclude); err != nil {
		suite.FailNow(err.Error())
	}

	// Get the exclude again.
	dbExclude, err := suite.state.DB.GetDomainPermissionExcludeByID(ctx, exclude.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Domain should have been stored punycoded.
	suite.Equal("xn--exmple-cua.org", dbExclude.Domain)

	// Search for domain using both
	// punycode and unicode variants.
	search1, err := suite.state.DB.GetDomainPermissionExcludes(
		ctx,
		"exämple.org",
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if len(search1) != 1 {
		suite.FailNow("couldn't get domain perm exclude exämple.org")
	}

	search2, err := suite.state.DB.GetDomainPermissionExcludes(
		ctx,
		"xn--exmple-cua.org",
		nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}
	if len(search2) != 1 {
		suite.FailNow("couldn't get domain perm exclude example.org")
	}

	// Change ID + try to put the same exclude again.
	exclude.ID = "01JCZNVYSDT3JE385FABMJ7ADQ"
	err = suite.state.DB.PutDomainPermissionExclude(ctx, exclude)
	if !errors.Is(err, db.ErrAlreadyExists) {
		suite.FailNow("was able to insert same domain perm exclude twice")
	}

	// Delete both excludes.
	for _, id := range []string{
		"01JCZN614XG85GCGAMSV9ZZAEJ",
		"01JCZNVYSDT3JE385FABMJ7ADQ",
	} {
		if err := suite.state.DB.DeleteDomainPermissionExclude(ctx, id); err != nil {
			suite.FailNow("error deleting domain permission exclude")
		}
	}
}

func TestDomainPermissionExcludeTestSuite(t *testing.T) {
	suite.Run(t, new(DomainPermissionExcludeTestSuite))
}
