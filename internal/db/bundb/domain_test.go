/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package bundb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type DomainTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DomainTestSuite) TestIsDomainBlocked() {
	ctx := context.Background()

	domainBlock := &gtsmodel.DomainBlock{
		ID:                 "01G204214Y9TNJEBX39C7G88SW",
		Domain:             "some.bad.apples",
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
		CreatedByAccount:   suite.testAccounts["admin_account"],
	}

	// no domain block exists for the given domain yet
	blocked, err := suite.db.IsDomainBlocked(ctx, domainBlock.Domain)
	suite.NoError(err)
	suite.False(blocked)

	err = suite.db.CreateDomainBlock(ctx, domainBlock)
	suite.NoError(err)

	// domain block now exists
	blocked, err = suite.db.IsDomainBlocked(ctx, domainBlock.Domain)
	suite.NoError(err)
	suite.True(blocked)
	suite.WithinDuration(time.Now(), domainBlock.CreatedAt, 10*time.Second)
}

func (suite *DomainTestSuite) TestIsDomainBlockedWildcard() {
	ctx := context.Background()

	domainBlock := &gtsmodel.DomainBlock{
		ID:                 "01G204214Y9TNJEBX39C7G88SW",
		Domain:             "bad.apples",
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
		CreatedByAccount:   suite.testAccounts["admin_account"],
	}

	// no domain block exists for the given domain yet
	blocked, err := suite.db.IsDomainBlocked(ctx, domainBlock.Domain)
	suite.NoError(err)
	suite.False(blocked)

	err = suite.db.CreateDomainBlock(ctx, domainBlock)
	suite.NoError(err)

	// Start with the base block domain
	domain := domainBlock.Domain

	for _, part := range []string{"extra", "domain", "parts"} {
		// Prepend the next domain part
		domain = part + "." + domain

		// Check that domain block is wildcarded for this subdomain
		blocked, err = suite.db.IsDomainBlocked(ctx, domainBlock.Domain)
		suite.NoError(err)
		suite.True(blocked)
	}
}

func (suite *DomainTestSuite) TestIsDomainBlockedNonASCII() {
	ctx := context.Background()

	now := time.Now()

	domainBlock := &gtsmodel.DomainBlock{
		ID:                 "01G204214Y9TNJEBX39C7G88SW",
		Domain:             "xn--80aaa1bbb1h.com",
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
		CreatedByAccount:   suite.testAccounts["admin_account"],
	}

	// no domain block exists for the given domain yet
	blocked, err := suite.db.IsDomainBlocked(ctx, "какашка.com")
	suite.NoError(err)
	suite.False(blocked)

	blocked, err = suite.db.IsDomainBlocked(ctx, "xn--80aaa1bbb1h.com")
	suite.NoError(err)
	suite.False(blocked)

	err = suite.db.CreateDomainBlock(ctx, domainBlock)
	suite.NoError(err)

	// domain block now exists
	blocked, err = suite.db.IsDomainBlocked(ctx, "какашка.com")
	suite.NoError(err)
	suite.True(blocked)

	blocked, err = suite.db.IsDomainBlocked(ctx, "xn--80aaa1bbb1h.com")
	suite.NoError(err)
	suite.True(blocked)
}

func (suite *DomainTestSuite) TestIsDomainBlockedNonASCII2() {
	ctx := context.Background()

	now := time.Now()

	domainBlock := &gtsmodel.DomainBlock{
		ID:                 "01G204214Y9TNJEBX39C7G88SW",
		Domain:             "какашка.com",
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedByAccountID: suite.testAccounts["admin_account"].ID,
		CreatedByAccount:   suite.testAccounts["admin_account"],
	}

	// no domain block exists for the given domain yet
	blocked, err := suite.db.IsDomainBlocked(ctx, "какашка.com")
	suite.NoError(err)
	suite.False(blocked)

	blocked, err = suite.db.IsDomainBlocked(ctx, "xn--80aaa1bbb1h.com")
	suite.NoError(err)
	suite.False(blocked)

	err = suite.db.CreateDomainBlock(ctx, domainBlock)
	suite.NoError(err)

	// domain block now exists
	blocked, err = suite.db.IsDomainBlocked(ctx, "какашка.com")
	suite.NoError(err)
	suite.True(blocked)

	blocked, err = suite.db.IsDomainBlocked(ctx, "xn--80aaa1bbb1h.com")
	suite.NoError(err)
	suite.True(blocked)
}

func TestDomainTestSuite(t *testing.T) {
	suite.Run(t, new(DomainTestSuite))
}
