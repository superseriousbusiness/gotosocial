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

package validate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func happyStatus() *gtsmodel.Status {
	return &gtsmodel.Status{
		ID:                       "01FEBBH6NYDG87NK6A6EC543ED",
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		URI:                      "https://example.org/users/test_user/statuses/01FEBBH6NYDG87NK6A6EC543ED",
		URL:                      "https://example.org/@test_user/01FEBBH6NYDG87NK6A6EC543ED",
		Content:                  "<p>Test status! #hello</p>",
		AttachmentIDs:            []string{"01FEBBKZBY9H5FEP3PHVVAAGN1", "01FEBBM7S2R4WT6WWW22KN1PWE"},
		Attachments:              nil,
		TagIDs:                   []string{"01FEBBNBMBSN1FESMZ1TCXNWYP"},
		Tags:                     nil,
		MentionIDs:               nil,
		Mentions:                 nil,
		EmojiIDs:                 nil,
		Emojis:                   nil,
		Local:                    testrig.TrueBool(),
		AccountID:                "01FEBBQ4KEP3824WW61MF52638",
		Account:                  nil,
		AccountURI:               "https://example.org/users/test_user",
		InReplyToID:              "",
		InReplyToURI:             "",
		InReplyToAccountID:       "",
		InReplyTo:                nil,
		InReplyToAccount:         nil,
		BoostOfID:                "",
		BoostOfAccountID:         "",
		BoostOf:                  nil,
		BoostOfAccount:           nil,
		ContentWarning:           "hello world test post",
		Visibility:               gtsmodel.VisibilityPublic,
		Sensitive:                testrig.FalseBool(),
		Language:                 "en",
		CreatedWithApplicationID: "01FEBBZHF4GFVRXSJVXD0JTZZ2",
		CreatedWithApplication:   nil,
		Federated:                testrig.TrueBool(),
		Boostable:                testrig.TrueBool(),
		Replyable:                testrig.TrueBool(),
		Likeable:                 testrig.TrueBool(),
		ActivityStreamsType:      ap.ObjectNote,
		Text:                     "Test status! #hello",
		Pinned:                   testrig.FalseBool(),
	}
}

type StatusValidateTestSuite struct {
	suite.Suite
}

func (suite *StatusValidateTestSuite) TestValidateStatusHappyPath() {
	// no problem here
	s := happyStatus()
	err := validate.Struct(s)
	suite.NoError(err)
}

func (suite *StatusValidateTestSuite) TestValidateStatusBadID() {
	s := happyStatus()

	s.ID = ""
	err := validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	s.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *StatusValidateTestSuite) TestValidateStatusAttachmentIDs() {
	s := happyStatus()

	s.AttachmentIDs[0] = ""
	err := validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.AttachmentIDs[0]' Error:Field validation for 'AttachmentIDs[0]' failed on the 'ulid' tag")

	s.AttachmentIDs[0] = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.AttachmentIDs[0]' Error:Field validation for 'AttachmentIDs[0]' failed on the 'ulid' tag")

	s.AttachmentIDs[1] = ""
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.AttachmentIDs[0]' Error:Field validation for 'AttachmentIDs[0]' failed on the 'ulid' tag\nKey: 'Status.AttachmentIDs[1]' Error:Field validation for 'AttachmentIDs[1]' failed on the 'ulid' tag")

	s.AttachmentIDs = []string{}
	err = validate.Struct(s)
	suite.NoError(err)

	s.AttachmentIDs = nil
	err = validate.Struct(s)
	suite.NoError(err)
}

func (suite *StatusValidateTestSuite) TestStatusApplicationID() {
	s := happyStatus()

	s.CreatedWithApplicationID = ""
	err := validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.CreatedWithApplicationID' Error:Field validation for 'CreatedWithApplicationID' failed on the 'required_if' tag")

	s.Local = testrig.FalseBool()
	err = validate.Struct(s)
	suite.NoError(err)
}

func (suite *StatusValidateTestSuite) TestValidateStatusReplyFields() {
	s := happyStatus()

	s.InReplyToAccountID = "01FEBCTP6DN7961PN81C3DVM4N                         "
	err := validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.InReplyToID' Error:Field validation for 'InReplyToID' failed on the 'required_with' tag\nKey: 'Status.InReplyToURI' Error:Field validation for 'InReplyToURI' failed on the 'required_with' tag\nKey: 'Status.InReplyToAccountID' Error:Field validation for 'InReplyToAccountID' failed on the 'ulid' tag")

	s.InReplyToAccountID = "01FEBCTP6DN7961PN81C3DVM4N"
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.InReplyToID' Error:Field validation for 'InReplyToID' failed on the 'required_with' tag\nKey: 'Status.InReplyToURI' Error:Field validation for 'InReplyToURI' failed on the 'required_with' tag")

	s.InReplyToURI = "https://example.org/users/mmbop/statuses/aaaaaaaa"
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.InReplyToID' Error:Field validation for 'InReplyToID' failed on the 'required_with' tag")

	s.InReplyToID = "not a valid ulid"
	err = validate.Struct(s)
	suite.EqualError(err, "Key: 'Status.InReplyToID' Error:Field validation for 'InReplyToID' failed on the 'ulid' tag")

	s.InReplyToID = "01FEBD07E72DEY6YB9K10ZA6ST"
	err = validate.Struct(s)
	suite.NoError(err)
}

func TestStatusValidateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusValidateTestSuite))
}
