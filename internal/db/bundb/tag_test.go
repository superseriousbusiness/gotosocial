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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/stretchr/testify/suite"
)

type TagTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *TagTestSuite) TestGetTag() {
	testTag := suite.testTags["welcome"]

	dbTag, err := suite.db.GetTag(context.Background(), testTag.ID)
	suite.NoError(err)
	suite.NotNil(dbTag)
	suite.Equal(testTag.ID, dbTag.ID)
}

func (suite *TagTestSuite) TestGetTagByName() {
	testTag := suite.testTags["welcome"]

	// Name is normalized when doing
	// selects from the db, so these
	// should all yield the same result.
	for _, name := range []string{
		"WELCOME",
		"welcome",
		"Welcome",
		"WELCoME ",
	} {
		dbTag, err := suite.db.GetTagByName(context.Background(), name)
		suite.NoError(err)
		suite.NotNil(dbTag)
		suite.Equal(testTag.ID, dbTag.ID)
	}
}

func (suite *TagTestSuite) TestPutTag() {
	// Name is normalized when doing
	// inserts to the db, so these
	// should all yield the same result.
	for i, name := range []string{
		"NewTag",
		"newtag",
		"NEWtag",
		"NEWTAG ",
	} {
		err := suite.db.PutTag(context.Background(), &gtsmodel.Tag{
			ID:   id.NewULID(),
			Name: name,
		})
		if i == 0 {
			// This is the first one, so it
			// should have just been created.
			suite.NoError(err)
			continue
		}

		// Subsequent inserts should fail
		// since all these tags are equivalent.
		if !suite.ErrorIs(err, db.ErrAlreadyExists) {
			suite.T().Logf("%T(%v) %v", err, err, unwrap(err))
		}
	}
}

func TestTagTestSuite(t *testing.T) {
	suite.Run(t, new(TagTestSuite))
}

func unwrap(err error) (errs []error) {
	for err != nil {
		errs = append(errs, err)
		err = errors.Unwrap(err)
	}
	return
}
