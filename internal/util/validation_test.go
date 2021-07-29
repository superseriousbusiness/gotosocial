/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package util_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type ValidationTestSuite struct {
	suite.Suite
}

func (suite *ValidationTestSuite) TestCheckPasswordStrength() {
	empty := ""
	terriblePassword := "password"
	weakPassword := "OKPassword"
	shortPassword := "Ok12"
	specialPassword := "Ok12%"
	longPassword := "thisisafuckinglongpasswordbutnospecialchars"
	tooLong := "Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Quisque a enim nibh. Vestibulum bibendum leo ac porttitor auctor."
	strongPassword := "3dX5@Zc%mV*W2MBNEy$@"
	var err error

	err = util.ValidateNewPassword(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no password provided"), err)
	}

	err = util.ValidateNewPassword(terriblePassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("insecure password, try including more special characters, using uppercase letters, using numbers or using a longer password"), err)
	}

	err = util.ValidateNewPassword(weakPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("insecure password, try including more special characters, using numbers or using a longer password"), err)
	}

	err = util.ValidateNewPassword(shortPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("insecure password, try including more special characters or using a longer password"), err)
	}

	err = util.ValidateNewPassword(specialPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("insecure password, try including more special characters or using a longer password"), err)
	}

	err = util.ValidateNewPassword(longPassword)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateNewPassword(tooLong)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password should be no more than 64 chars"), err)
	}

	err = util.ValidateNewPassword(strongPassword)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateUsername() {
	empty := ""
	tooLong := "holycrapthisisthelongestusernameiveeverseeninmylifethatstoomuchman"
	withSpaces := "this username has spaces in it"
	weirdChars := "thisusername&&&&&&&istooweird!!"
	leadingSpace := " see_that_leading_space"
	trailingSpace := "thisusername_ends_with_a_space "
	newlines := "this_is\n_almost_ok"
	goodUsername := "this_is_a_good_username"
	var err error

	err = util.ValidateUsername(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no username provided"), err)
	}

	err = util.ValidateUsername(tooLong)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", tooLong), err)
	}

	err = util.ValidateUsername(withSpaces)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", withSpaces), err)
	}

	err = util.ValidateUsername(weirdChars)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", weirdChars), err)
	}

	err = util.ValidateUsername(leadingSpace)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", leadingSpace), err)
	}

	err = util.ValidateUsername(trailingSpace)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", trailingSpace), err)
	}

	err = util.ValidateUsername(newlines)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", newlines), err)
	}

	err = util.ValidateUsername(goodUsername)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateEmail() {
	empty := ""
	notAnEmailAddress := "this-is-no-email-address!"
	almostAnEmailAddress := "@thisisalmostan@email.address"
	aWebsite := "https://thisisawebsite.com"
	emailAddress := "thisis.actually@anemail.address"
	var err error

	err = util.ValidateEmail(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no email provided"), err)
	}

	err = util.ValidateEmail(notAnEmailAddress)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = util.ValidateEmail(almostAnEmailAddress)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: no angle-addr"), err)
	}

	err = util.ValidateEmail(aWebsite)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = util.ValidateEmail(emailAddress)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateLanguage() {
	empty := ""
	notALanguage := "this isn't a language at all!"
	english := "en"
	capitalEnglish := "EN"
	arabic3Letters := "ara"
	mixedCapsEnglish := "eN"
	englishUS := "en-us"
	dutch := "nl"
	german := "de"
	var err error

	err = util.ValidateLanguage(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no language provided"), err)
	}

	err = util.ValidateLanguage(notALanguage)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("language: tag is not well-formed"), err)
	}

	err = util.ValidateLanguage(english)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateLanguage(capitalEnglish)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateLanguage(arabic3Letters)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateLanguage(mixedCapsEnglish)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateLanguage(englishUS)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("language: tag is not well-formed"), err)
	}

	err = util.ValidateLanguage(dutch)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateLanguage(german)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateReason() {
	empty := ""
	badReason := "because"
	goodReason := "to smash the state and destroy capitalism ultimately and completely"
	tooLong := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris auctor mollis viverra. Maecenas maximus mollis sem, nec fermentum velit consectetur non. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Quisque a enim nibh. Vestibulum bibendum leo ac porttitor auctor. Curabitur velit tellus, facilisis vitae lorem a, ullamcorper efficitur leo. Sed a auctor tortor. Sed ut finibus ante, sit amet laoreet sapien. Donec ullamcorper tellus a nibh sodales vulputate. Donec id dolor eu odio mollis bibendum. Pellentesque habitant morbi tristique senectus et netus at."
	var err error

	// check with no reason required
	err = util.ValidateSignUpReason(empty, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateSignUpReason(badReason, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateSignUpReason(tooLong, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = util.ValidateSignUpReason(goodReason, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	// check with reason required
	err = util.ValidateSignUpReason(empty, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no reason provided"), err)
	}

	err = util.ValidateSignUpReason(badReason, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("reason should be at least 40 chars but 'because' was 7"), err)
	}

	err = util.ValidateSignUpReason(tooLong, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("reason should be no more than 500 chars but given reason was 600"), err)
	}

	err = util.ValidateSignUpReason(goodReason, true)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
