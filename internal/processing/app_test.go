/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package processing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type AppTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *AppTestSuite) TestAppCreateWithNewlineSeparatedRedirectURIs() {
	ctx := context.Background()

	authed := suite.testAutheds["local_account_1"]

	app, errWithCode := suite.processor.AppCreate(ctx, authed, &apimodel.ApplicationCreateRequest{
		ClientName:   "abcde",
		RedirectURIs: "urn:ietf:wg:oauth:2.0:oob\nhttp://example.com",
		Scopes:       "read write follow push",
	})

	suite.NoError(errWithCode)

	suite.Equal(app.RedirectURI, "http://example.com")
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, &AccountTestSuite{})
}
