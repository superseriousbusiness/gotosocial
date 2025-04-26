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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
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

func (suite *DomainPermissionExcludeTestSuite) TestExcluded() {
	var (
		ctx                = context.Background()
		createdByAccountID = suite.testAccounts["admin_account"].ID
	)

	// Insert some excludes into the db.
	for _, exclude := range []*gtsmodel.DomainPermissionExclude{
		{
			ID:                 "01JD7AFFBBZSPY8R2M0JCGQGPW",
			Domain:             "example.org",
			CreatedByAccountID: createdByAccountID,
		},
		{
			ID:                 "01JD7AMK98E2QX78KXEZJ1RF5Z",
			Domain:             "boobs.com",
			CreatedByAccountID: createdByAccountID,
		},
		{
			ID:                 "01JD7AMXW3R3W98E91R62ACDA0",
			Domain:             "rad.boobs.com",
			CreatedByAccountID: createdByAccountID,
		},
		{
			ID:                 "01JD7AYYN5TXQVASB30PT08CE1",
			Domain:             "honkers.org",
			CreatedByAccountID: createdByAccountID,
		},
	} {
		if err := suite.state.DB.PutDomainPermissionExclude(ctx, exclude); err != nil {
			suite.FailNow(err.Error())
		}
	}

	type testCase struct {
		domain   string
		excluded bool
	}

	for i, testCase := range []testCase{
		{
			domain:   config.GetHost(),
			excluded: true,
		},
		{
			domain:   "test.example.org",
			excluded: true,
		},
		{
			domain:   "example.org",
			excluded: true,
		},
		{
			domain:   "boobs.com",
			excluded: true,
		},
		{
			domain:   "rad.boobs.com",
			excluded: true,
		},
		{
			domain:   "sir.not.appearing.in.this.list",
			excluded: false,
		},
	} {
		excluded, err := suite.state.DB.IsDomainPermissionExcluded(ctx, testCase.domain)
		if err != nil {
			suite.FailNow(err.Error())
		}

		if excluded != testCase.excluded {
			suite.Failf("",
				"test %d: %s excluded should be %t",
				i, testCase.domain, testCase.excluded,
			)
		}
	}
}

func TestDomainPermissionExcludeTestSuite(t *testing.T) {
	suite.Run(t, new(DomainPermissionExcludeTestSuite))
}
