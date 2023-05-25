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

package validate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	tooLong := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed euismod, ante id iaculis suscipit, nibh nibh varius enim, eget euismod augue augue eget mi. Praesent tincidunt, ex id finibus congue, enim nunc euismod nulla, id tincidunt ipsum neque at nunc. Sed id convallis libero. Sed euismod augue augue eget mi. Praesent tincidunt, ex id finibus congue, enim nunc euismod nulla, id tincidunt ipsum neque at nunc. Sed id convallis libero. Sed euismod augue augue eget mi. Praesent tincidunt, ex id finibus congue, enim nunc euismod nulla, id tincidunt ipsum neque at nunc."
	strongPassword := "3dX5@Zc%mV*W2MBNEy$@"
	var err error

	err = validate.NewPassword(empty)
	if suite.Error(err) {
		suite.Equal(errors.New("no password provided"), err)
	}

	err = validate.NewPassword(terriblePassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 62% strength, try including more special characters, using uppercase letters, using numbers or using a longer password"), err)
	}

	err = validate.NewPassword(weakPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 95% strength, try including more special characters, using numbers or using a longer password"), err)
	}

	err = validate.NewPassword(shortPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 39% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.NewPassword(specialPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 53% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.NewPassword(longPassword)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.NewPassword(tooLong)
	if suite.Error(err) {
		suite.Equal(errors.New("password should be no more than 256 chars"), err)
	}

	err = validate.NewPassword(strongPassword)
	if suite.NoError(err) {
		suite.Equal(nil, err)
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
	singleChar := "s"
	var err error

	err = validate.Username(empty)
	suite.EqualError(err, "no username provided")

	err = validate.Username(tooLong)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", tooLong))

	err = validate.Username(withSpaces)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", withSpaces))

	err = validate.Username(weirdChars)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", weirdChars))

	err = validate.Username(leadingSpace)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", leadingSpace))

	err = validate.Username(trailingSpace)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", trailingSpace))

	err = validate.Username(newlines)
	suite.EqualError(err, fmt.Sprintf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max 64 characters", newlines))

	err = validate.Username(goodUsername)
	suite.NoError(err)

	err = validate.Username(singleChar)
	suite.NoError(err)
}

func (suite *ValidationTestSuite) TestValidateEmail() {
	empty := ""
	notAnEmailAddress := "this-is-no-email-address!"
	almostAnEmailAddress := "@thisisalmostan@email.address"
	aWebsite := "https://thisisawebsite.com"
	emailAddress := "thisis.actually@anemail.address"
	var err error

	err = validate.Email(empty)
	if suite.Error(err) {
		suite.Equal(errors.New("no email provided"), err)
	}

	err = validate.Email(notAnEmailAddress)
	if suite.Error(err) {
		suite.Equal(errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = validate.Email(almostAnEmailAddress)
	if suite.Error(err) {
		suite.Equal(errors.New("mail: no angle-addr"), err)
	}

	err = validate.Email(aWebsite)
	if suite.Error(err) {
		suite.Equal(errors.New("mail: missing '@' or angle-addr"), err)
	}

	err = validate.Email(emailAddress)
	if suite.NoError(err) {
		suite.Equal(nil, err)
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
	if suite.Error(err) {
		suite.Equal(errors.New("no language provided"), err)
	}

	err = validate.Language(notALanguage)
	if suite.Error(err) {
		suite.Equal(errors.New("language: tag is not well-formed"), err)
	}

	err = validate.Language(english)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Language(capitalEnglish)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Language(arabic3Letters)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Language(mixedCapsEnglish)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Language(englishUS)
	if suite.Error(err) {
		suite.Equal(errors.New("language: tag is not well-formed"), err)
	}

	err = validate.Language(dutch)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Language(german)
	if suite.NoError(err) {
		suite.Equal(nil, err)
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
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.SignUpReason(badReason, false)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.SignUpReason(tooLong, false)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.SignUpReason(goodReason, false)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.SignUpReason(unicode, false)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	// check with reason required
	err = validate.SignUpReason(empty, true)
	if suite.Error(err) {
		suite.Equal(errors.New("no reason provided"), err)
	}

	err = validate.SignUpReason(badReason, true)
	if suite.Error(err) {
		suite.Equal(errors.New("reason should be at least 40 chars but 'because' was 7"), err)
	}

	err = validate.SignUpReason(tooLong, true)
	if suite.Error(err) {
		suite.Equal(errors.New("reason should be no more than 500 chars but given reason was 600"), err)
	}

	err = validate.SignUpReason(goodReason, true)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}
}

func (suite *ValidationTestSuite) TestValidateProfileField() {
	var (
		shortProfileField   = "pronouns"
		tooLongProfileField = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Integer eu bibendum elit. Sed ac interdum nisi. Vestibulum vulputate eros quis euismod imperdiet. Nulla sit amet dui sit amet lorem consectetur iaculis. Mauris eget lacinia metus. Curabitur nec dui eleifend massa nunc."
		trimmedProfileField = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Integer eu bibendum elit. Sed ac interdum nisi. Vestibulum vulputate eros quis euismod imperdiet. Nulla sit amet dui sit amet lorem consectetur iaculis. Mauris eget lacinia metus. Curabitur nec dui "
		err                 error
	)

	okFields := []*gtsmodel.Field{
		{
			Name:  "example",
			Value: shortProfileField,
		},
	}
	err = validate.ProfileFields(okFields)
	suite.NoError(err)
	suite.Equal(shortProfileField, okFields[0].Value)

	dodgyFields := []*gtsmodel.Field{
		{
			Name:  "example",
			Value: tooLongProfileField,
		},
	}
	err = validate.ProfileFields(dodgyFields)
	suite.NoError(err)
	suite.Equal(trimmedProfileField, dodgyFields[0].Value)
	suite.Len(dodgyFields[0].Value, 255)
}

func (suite *ValidationTestSuite) TestValidateCustomCSSDisabled() {
	config.SetAccountsAllowCustomCSS(false)

	err := validate.CustomCSS("this will fail")
	suite.EqualError(err, "accounts-allow-custom-css is not enabled for this instance")
}

func (suite *ValidationTestSuite) TestValidateCustomCSSEnabled() {
	config.SetAccountsAllowCustomCSS(true)

	err := validate.CustomCSS("this will pass")
	suite.NoError(err)
}

func (suite *ValidationTestSuite) TestValidateCustomCSSTooLong() {
	config.SetAccountsAllowCustomCSS(true)
	config.SetAccountsCustomCSSLength(5)

	err := validate.CustomCSS("this will fail")
	suite.EqualError(err, "custom_css must be less than 5 characters, but submitted custom_css was 14 characters")
}

func (suite *ValidationTestSuite) TestValidateCustomCSSTooLongZalgo() {
	config.SetAccountsAllowCustomCSS(true)
	config.SetAccountsCustomCSSLength(5)
	zalgo := "p̵̹̜͇̺̜̱͊̓̈́͛̀͊͘͜e̷̡̱̲̼̪̗̙̐͐̃́̄̉͛̔e̷̞̰̜̲̥̘̻͔̜̞̬͚͋̊͑͗̅̓͛͗̎̃̈́̐̂̕͝ ̷̨̢̡̱̖̤͇̻͕̲̤̞̑ͅp̶̰̜̟̠̏̇̇̆̐̒͋̔͘ḛ̵̾͘ę̷̝͙͕͓͓̱̠̤̳̻̜̗͖̞͙̻̆̓̄͋̎͊̀̋̿́̐͛͗̄̈́̚͠ ̵̨̨̫͕̲͚̮͕̳̉̾̔̍͐p̶̘̞̠̘̎̓̍̑̀͗̃̈́͂́̈́͆͘͜͝͝o̶̜͛̒͒̉̑͒̿͗̐̃͝o̵̼̒͌̓ ̵̢̗̦͔͉͈̰̘̋̃̐̑̅̽̏̄̅͐͆̔͊̃̋͝p̵̩̱̆̆͂̂͛̓̋̅͝o̶̪̰̲̝̻̳̦̮̮͔̒ͅơ̸̧̨̟͇̪̰̜̠̦͇̇̎͗̏̏̈́͂̉̏͐́̃̀͆͠ͅ"

	err := validate.CustomCSS(zalgo)
	suite.EqualError(err, "custom_css must be less than 5 characters, but submitted custom_css was 275 characters")
}

func (suite *ValidationTestSuite) TestValidateCustomCSSTooLongUnicode() {
	config.SetAccountsAllowCustomCSS(true)
	config.SetAccountsCustomCSSLength(5)
	unicode := "⎾⎿⏀⏁⏂⏃⏄⏅⏆⏇"

	err := validate.CustomCSS(unicode)
	suite.EqualError(err, "custom_css must be less than 5 characters, but submitted custom_css was 10 characters")
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
