package gtsmodel_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func happyUser() *gtsmodel.User {
	return &gtsmodel.User{
		ID:                     "01FE8TTK9F34BR0KG7639AJQTX",
		Email:                  "whatever@example.org",
		AccountID:              "01FE8TWA7CN8J7237K5DFS1RY5",
		Account:                nil,
		EncryptedPassword:      "$2y$10$tkRapNGW.RWkEuCMWdgArunABFvsPGRvFQY3OibfSJo0RDL3z8WfC",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		SignUpIP:               net.ParseIP("128.64.32.16"),
		CurrentSignInAt:        time.Now(),
		CurrentSignInIP:        net.ParseIP("128.64.32.16"),
		LastSignInAt:           time.Now(),
		LastSignInIP:           net.ParseIP("128.64.32.16"),
		SignInCount:            0,
		InviteID:               "",
		ChosenLanguages:        []string{},
		FilteredLanguages:      []string{},
		Locale:                 "en",
		CreatedByApplicationID: "01FE8Y5EHMWCA1MHMTNHRVZ1X4",
		CreatedByApplication:   nil,
		LastEmailedAt:          time.Now(),
		ConfirmationToken:      "",
		ConfirmedAt:            time.Now(),
		ConfirmationSentAt:     time.Now(),
		UnconfirmedEmail:       "",
		Moderator:              false,
		Admin:                  false,
		Disabled:               false,
		Approved:               true,
	}
}

type UserValidateTestSuite struct {
	suite.Suite
}

func (suite *UserValidateTestSuite) TestValidateUserHappyPath() {
	// no problem here
	u := happyUser()
	err := gtsmodel.ValidateStruct(*u)
	suite.NoError(err)
}

func (suite *UserValidateTestSuite) TestValidateUserNoID() {
	// user has no id set
	u := happyUser()
	u.ID = ""

	err := gtsmodel.ValidateStruct(*u)
	suite.EqualError(err, "Key: 'User.ID' Error:Field validation for 'ID' failed on the 'required' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserNoEmail() {
	// user has no email or unconfirmed email set
	u := happyUser()
	u.Email = ""

	err := gtsmodel.ValidateStruct(*u)
	suite.EqualError(err, "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required_with' tag\nKey: 'User.UnconfirmedEmail' Error:Field validation for 'UnconfirmedEmail' failed on the 'required_without' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserOnlyUnconfirmedEmail() {
	// user has only UnconfirmedEmail but ConfirmedAt is set
	u := happyUser()
	u.Email = ""
	u.UnconfirmedEmail = "whatever@example.org"

	err := gtsmodel.ValidateStruct(*u)
	suite.EqualError(err, "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required_with' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserOnlyUnconfirmedEmailOK() {
	// user has only UnconfirmedEmail and ConfirmedAt is not set
	u := happyUser()
	u.Email = ""
	u.UnconfirmedEmail = "whatever@example.org"
	u.ConfirmedAt = time.Time{}

	err := gtsmodel.ValidateStruct(*u)
	suite.NoError(err)
}

func (suite *UserValidateTestSuite) TestValidateUserNoConfirmedAt() {
	// user has Email but no ConfirmedAt
	u := happyUser()
	u.ConfirmedAt = time.Time{}

	err := gtsmodel.ValidateStruct(*u)
	suite.EqualError(err, "Key: 'User.ConfirmedAt' Error:Field validation for 'ConfirmedAt' failed on the 'required_with' tag")
}

func TestUserValidateTestSuite(t *testing.T) {
	suite.Run(t, new(UserValidateTestSuite))
}
