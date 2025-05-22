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

package account_test

import (
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type AccountUpdateTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateSimple() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	var (
		ctx          = suite.T().Context()
		locked       = true
		displayName  = "new display name"
		note         = "#hello here i am!"
		noteExpected = `<p><a href="http://localhost:8080/tags/hello" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>hello</span></a> here i am!</p>`
	)

	// Call update function.
	apiAccount, errWithCode := suite.accountProcessor.Update(ctx, testAccount, &apimodel.UpdateCredentialsRequest{
		DisplayName: &displayName,
		Locked:      &locked,
		Note:        &note,
	})
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.True(apiAccount.Locked)
	suite.Equal(displayName, apiAccount.DisplayName)
	suite.Equal(noteExpected, apiAccount.Note)

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)

	// Profile update.
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ActorPerson, msg.APObjectType)

	// Correct account updated.
	if msg.Origin == nil {
		suite.FailNow("expected msg.OriginAccount not to be nil")
	}
	suite.Equal(testAccount.ID, msg.Origin.ID)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(*dbAccount.Locked)
	suite.Equal(displayName, dbAccount.DisplayName)
	suite.Equal(noteExpected, dbAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateWithMention() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	var (
		ctx          = suite.T().Context()
		locked       = true
		displayName  = "new display name"
		note         = "#hello here i am!\n\ngo check out @1happyturtle, they have a cool account!"
		noteExpected = "<p><a href=\"http://localhost:8080/tags/hello\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hello</span></a> here i am!<br><br>go check out <span class=\"h-card\"><a href=\"http://localhost:8080/@1happyturtle\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>1happyturtle</span></a></span>, they have a cool account!</p>"
	)

	// Call update function.
	apiAccount, errWithCode := suite.accountProcessor.Update(ctx, testAccount, &apimodel.UpdateCredentialsRequest{
		DisplayName: &displayName,
		Locked:      &locked,
		Note:        &note,
	})
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.True(apiAccount.Locked)
	suite.Equal(displayName, apiAccount.DisplayName)
	suite.Equal(noteExpected, apiAccount.Note)

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)

	// Profile update.
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ActorPerson, msg.APObjectType)

	// Correct account updated.
	if msg.Origin == nil {
		suite.FailNow("expected msg.OriginAccount not to be nil")
	}
	suite.Equal(testAccount.ID, msg.Origin.ID)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(*dbAccount.Locked)
	suite.Equal(displayName, dbAccount.DisplayName)
	suite.Equal(noteExpected, dbAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateWithMarkdownNote() {
	// Copy zork.
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	// Copy zork's settings.
	settings := &gtsmodel.AccountSettings{}
	*settings = *suite.testAccounts["local_account_1"].Settings
	testAccount.Settings = settings

	var (
		ctx          = suite.T().Context()
		note         = "*hello* ~~here~~ i am!"
		noteExpected = `<p><em>hello</em> <del>here</del> i am!</p>`
	)

	// Set status content type of account 1 to markdown for this test.
	testAccount.Settings.StatusContentType = "text/markdown"
	if err := suite.db.UpdateAccountSettings(ctx, testAccount.Settings, "status_content_type"); err != nil {
		suite.FailNow(err.Error())
	}

	// Call update function.
	apiAccount, errWithCode := suite.accountProcessor.Update(ctx, testAccount, &apimodel.UpdateCredentialsRequest{
		Note: &note,
	})
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.Equal(noteExpected, apiAccount.Note)

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)

	// Profile update.
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ActorPerson, msg.APObjectType)

	// Correct account updated.
	if msg.Origin == nil {
		suite.FailNow("expected msg.OriginAccount not to be nil")
	}
	suite.Equal(testAccount.ID, msg.Origin.ID)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NoError(err)
	suite.Equal(noteExpected, dbAccount.Note)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateWithFields() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]

	var (
		ctx          = suite.T().Context()
		updateFields = []apimodel.UpdateField{
			{
				Name:  func() *string { s := "favourite emoji"; return &s }(),
				Value: func() *string { s := ":rainbow:"; return &s }(),
			},
			{
				Name:  func() *string { s := "my website"; return &s }(),
				Value: func() *string { s := "https://example.org"; return &s }(),
			},
		}
		fieldsExpectedRaw = []apimodel.Field{
			{
				Name:       "favourite emoji",
				Value:      ":rainbow:",
				VerifiedAt: (*string)(nil),
			},
			{
				Name:       "my website",
				Value:      "https://example.org",
				VerifiedAt: (*string)(nil),
			},
		}
		fieldsExpectedParsed = []apimodel.Field{
			{
				Name:       "favourite emoji",
				Value:      ":rainbow:",
				VerifiedAt: (*string)(nil),
			},
			{
				Name:       "my website",
				Value:      "<a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://example.org</a>",
				VerifiedAt: (*string)(nil),
			},
		}
		emojisExpected = []apimodel.Emoji{
			{
				Shortcode:       "rainbow",
				URL:             "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/original/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
				StaticURL:       "http://localhost:8080/fileserver/01AY6P665V14JJR0AFVRT7311Y/emoji/static/01F8MH9H8E4VG3KDYJR9EGPXCQ.png",
				VisibleInPicker: true,
				Category:        "reactions",
			},
		}
	)

	apiAccount, errWithCode := suite.accountProcessor.Update(ctx, testAccount, &apimodel.UpdateCredentialsRequest{
		FieldsAttributes: &updateFields,
	})
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.EqualValues(fieldsExpectedRaw, apiAccount.Source.Fields)
	suite.EqualValues(fieldsExpectedParsed, apiAccount.Fields)
	suite.EqualValues(emojisExpected, apiAccount.Emojis)

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)

	// Profile update.
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ActorPerson, msg.APObjectType)

	// Correct account updated.
	if msg.Origin == nil {
		suite.FailNow("expected msg.OriginAccount not to be nil")
	}
	suite.Equal(testAccount.ID, msg.Origin.ID)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(fieldsExpectedParsed[0].Name, dbAccount.Fields[0].Name)
	suite.Equal(fieldsExpectedParsed[0].Value, dbAccount.Fields[0].Value)
	suite.Equal(fieldsExpectedParsed[1].Name, dbAccount.Fields[1].Name)
	suite.Equal(fieldsExpectedParsed[1].Value, dbAccount.Fields[1].Value)
	suite.Equal(fieldsExpectedRaw[0].Name, dbAccount.FieldsRaw[0].Name)
	suite.Equal(fieldsExpectedRaw[0].Value, dbAccount.FieldsRaw[0].Value)
	suite.Equal(fieldsExpectedRaw[1].Name, dbAccount.FieldsRaw[1].Name)
	suite.Equal(fieldsExpectedRaw[1].Value, dbAccount.FieldsRaw[1].Value)
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateNoteNotFields() {
	// local_account_2 already has some fields set.
	// We want to ensure that the fields don't change
	// even if the account note is updated.
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_2"]

	var (
		ctx             = suite.T().Context()
		fieldsRawBefore = len(testAccount.FieldsRaw)
		fieldsBefore    = len(testAccount.Fields)
		note            = "#hello here i am!"
		noteExpected    = `<p><a href="http://localhost:8080/tags/hello" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>hello</span></a> here i am!</p>`
	)

	// Call update function.
	apiAccount, errWithCode := suite.accountProcessor.Update(ctx, testAccount, &apimodel.UpdateCredentialsRequest{
		Note: &note,
	})
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.True(apiAccount.Locked)
	suite.Equal(noteExpected, apiAccount.Note)
	suite.Equal(fieldsRawBefore, len(apiAccount.Source.Fields))
	suite.Equal(fieldsBefore, len(apiAccount.Fields))

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)

	// Profile update.
	suite.Equal(ap.ActivityUpdate, msg.APActivityType)
	suite.Equal(ap.ActorPerson, msg.APObjectType)

	// Correct account updated.
	if msg.Origin == nil {
		suite.FailNow("expected msg.OriginAccount not to be nil")
	}
	suite.Equal(testAccount.ID, msg.Origin.ID)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(*dbAccount.Locked)
	suite.Equal(noteExpected, dbAccount.Note)
	suite.Equal(fieldsRawBefore, len(dbAccount.FieldsRaw))
	suite.Equal(fieldsBefore, len(dbAccount.Fields))
}

func (suite *AccountUpdateTestSuite) TestAccountUpdateBotNotBot() {
	testAccount := &gtsmodel.Account{}
	*testAccount = *suite.testAccounts["local_account_1"]
	ctx := suite.T().Context()

	// Call update function to set bot = true.
	apiAccount, errWithCode := suite.accountProcessor.Update(
		ctx,
		testAccount,
		&apimodel.UpdateCredentialsRequest{
			Bot: util.Ptr(true),
		},
	)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.True(apiAccount.Bot)

	// We should have an update in the client api channel.
	msg, _ := suite.getClientMsg(5 * time.Second)
	suite.NotNil(msg)

	// Check database model of account as well.
	dbAccount, err := suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(dbAccount.ActorType.IsBot())

	// Call update function to set bot = false.
	apiAccount, errWithCode = suite.accountProcessor.Update(
		ctx,
		testAccount,
		&apimodel.UpdateCredentialsRequest{
			Bot: util.Ptr(false),
		},
	)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Returned profile should be updated.
	suite.False(apiAccount.Bot)

	// We should have an update in the client api channel.
	msg, _ = suite.getClientMsg(5 * time.Second)
	suite.NotNil(msg)

	// Check database model of account as well.
	dbAccount, err = suite.db.GetAccountByID(ctx, testAccount.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.False(dbAccount.ActorType.IsBot())
}

func TestAccountUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(AccountUpdateTestSuite))
}
