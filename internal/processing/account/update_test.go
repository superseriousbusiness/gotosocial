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

package account_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateSimple() {
	testAccount := suite.testAccounts["local_account_1"]

	locked := true
	displayName := "new display name"
	note := "#hello here i am!"

	form := &apimodel.UpdateCredentialsRequest{
		DisplayName: &displayName,
		Locked:      &locked,
		Note:        &note,
	}

	// should get no error from the update function, and an api model account returned
	apiAccount, errWithCode := suite.accountProcessor.Update(context.Background(), testAccount, form)
	suite.NoError(errWithCode)
	suite.NotNil(apiAccount)

	// fields on the profile should be updated
	suite.True(apiAccount.Locked)
	suite.Equal(displayName, apiAccount.DisplayName)
	suite.Equal(`<p><a href="http://localhost:8080/tags/hello" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>hello</span></a> here i am!</p>`, apiAccount.Note)

	// we should have an update in the client api channel
	msg := <-suite.fromClientAPIChan
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ObjectProfile, msg.APObjectType)
	suite.NotNil(msg.OriginAccount)
	suite.Equal(testAccount.ID, msg.OriginAccount.ID)
	suite.Nil(msg.TargetAccount)

	// fields should be updated in the database as well
	dbAccount, err := suite.db.GetAccountByID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.True(*dbAccount.Locked)
	suite.Equal(displayName, dbAccount.DisplayName)
	suite.Equal(`<p><a href="http://localhost:8080/tags/hello" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>hello</span></a> here i am!</p>`, dbAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateWithMention() {
	testAccount := suite.testAccounts["local_account_1"]

	var (
		locked       = true
		displayName  = "new display name"
		note         = "#hello here i am!\n\ngo check out @1happyturtle, they have a cool account!\n"
		noteExpected = "<p><a href=\"http://localhost:8080/tags/hello\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hello</span></a> here i am!<br/><br/>go check out <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span>, they have a cool account!</p>"
	)

	form := &apimodel.UpdateCredentialsRequest{
		DisplayName: &displayName,
		Locked:      &locked,
		Note:        &note,
	}

	// should get no error from the update function, and an api model account returned
	apiAccount, errWithCode := suite.accountProcessor.Update(context.Background(), testAccount, form)
	suite.NoError(errWithCode)
	suite.NotNil(apiAccount)

	// fields on the profile should be updated
	suite.True(apiAccount.Locked)
	suite.Equal(displayName, apiAccount.DisplayName)
	suite.Equal(noteExpected, apiAccount.Note)

	// we should have an update in the client api channel
	msg := <-suite.fromClientAPIChan
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ObjectProfile, msg.APObjectType)
	suite.NotNil(msg.OriginAccount)
	suite.Equal(testAccount.ID, msg.OriginAccount.ID)
	suite.Nil(msg.TargetAccount)

	// fields should be updated in the database as well
	dbAccount, err := suite.db.GetAccountByID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.True(*dbAccount.Locked)
	suite.Equal(displayName, dbAccount.DisplayName)
	suite.Equal(noteExpected, dbAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateWithMarkdownNote() {
	testAccount := suite.testAccounts["local_account_1"]

	note := "*hello* ~~here~~ i am!"
	expectedNote := `<p><em>hello</em> <del>here</del> i am!</p>`

	form := &apimodel.UpdateCredentialsRequest{
		Note: &note,
	}

	// set default post language of account 1 to markdown
	testAccount.StatusFormat = "markdown"

	// should get no error from the update function, and an api model account returned
	apiAccount, errWithCode := suite.accountProcessor.Update(context.Background(), testAccount, form)
	// reset test account to avoid breaking other tests
	testAccount.StatusFormat = "plain"
	suite.NoError(errWithCode)
	suite.NotNil(apiAccount)

	// fields on the profile should be updated
	suite.Equal(expectedNote, apiAccount.Note)

	// we should have an update in the client api channel
	msg := <-suite.fromClientAPIChan
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ObjectProfile, msg.APObjectType)
	suite.NotNil(msg.OriginAccount)
	suite.Equal(testAccount.ID, msg.OriginAccount.ID)
	suite.Nil(msg.TargetAccount)

	// fields should be updated in the database as well
	dbAccount, err := suite.db.GetAccountByID(context.Background(), testAccount.ID)
	suite.NoError(err)
	suite.Equal(expectedNote, dbAccount.Note)

}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
