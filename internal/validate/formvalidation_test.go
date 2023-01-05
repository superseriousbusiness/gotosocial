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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
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

	err = validate.NewPassword(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no password provided"), err)
	}

	err = validate.NewPassword(terriblePassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password is only 62% strength, try including more special characters, using uppercase letters, using numbers or using a longer password"), err)
	}

	err = validate.NewPassword(weakPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password is only 95% strength, try including more special characters, using numbers or using a longer password"), err)
	}

	err = validate.NewPassword(shortPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password is only 39% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.NewPassword(specialPassword)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password is only 53% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.NewPassword(longPassword)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.NewPassword(tooLong)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("password should be no more than 64 chars"), err)
	}

	err = validate.NewPassword(strongPassword)
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

	err = validate.Username(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no username provided"), err)
	}

	err = validate.Username(tooLong)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", tooLong), err)
	}

	err = validate.Username(withSpaces)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", withSpaces), err)
	}

	err = validate.Username(weirdChars)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", weirdChars), err)
	}

	err = validate.Username(leadingSpace)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", leadingSpace), err)
	}

	err = validate.Username(trailingSpace)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", trailingSpace), err)
	}

	err = validate.Username(newlines)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", newlines), err)
	}

	err = validate.Username(goodUsername)
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

	err = validate.Email(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no email provided"), err)
	}

	err = validate.Email(notAnEmailAddress)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = validate.Email(almostAnEmailAddress)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: no angle-addr"), err)
	}

	err = validate.Email(aWebsite)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = validate.Email(emailAddress)
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

	err = validate.Language(empty)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no language provided"), err)
	}

	err = validate.Language(notALanguage)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("language: tag is not well-formed"), err)
	}

	err = validate.Language(english)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.Language(capitalEnglish)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.Language(arabic3Letters)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.Language(mixedCapsEnglish)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.Language(englishUS)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("language: tag is not well-formed"), err)
	}

	err = validate.Language(dutch)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.Language(german)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateReason() {
	empty := ""
	badReason := "because"
	goodReason := "to smash the state and destroy capitalism ultimately and completely"
	tooLong := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Mauris auctor mollis viverra. Maecenas maximus mollis sem, nec fermentum velit consectetur non. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Quisque a enim nibh. Vestibulum bibendum leo ac porttitor auctor. Curabitur velit tellus, facilisis vitae lorem a, ullamcorper efficitur leo. Sed a auctor tortor. Sed ut finibus ante, sit amet laoreet sapien. Donec ullamcorper tellus a nibh sodales vulputate. Donec id dolor eu odio mollis bibendum. Pellentesque habitant morbi tristique senectus et netus at."
	unicode := "⎾⎿⏀⏁⏂⏃⏄⏅⏆⏇"
	var err error

	// check with no reason required
	err = validate.SignUpReason(empty, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.SignUpReason(badReason, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.SignUpReason(tooLong, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.SignUpReason(goodReason, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	err = validate.SignUpReason(unicode, false)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}

	// check with reason required
	err = validate.SignUpReason(empty, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("no reason provided"), err)
	}

	err = validate.SignUpReason(badReason, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("reason should be at least 40 chars but 'because' was 7"), err)
	}

	err = validate.SignUpReason(tooLong, true)
	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), errors.New("reason should be no more than 500 chars but given reason was 600"), err)
	}

	err = validate.SignUpReason(goodReason, true)
	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), nil, err)
	}
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
