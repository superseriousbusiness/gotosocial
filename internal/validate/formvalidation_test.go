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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"github.com/stretchr/testify/suite"
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

	err = validate.Password(empty)
	if suite.Error(err) {
		suite.Equal(errors.New("no password provided / provided password was 0 bytes"), err)
	}

	err = validate.Password(terriblePassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 62% strength, try including more special characters, using uppercase letters, using numbers or using a longer password"), err)
	}

	err = validate.Password(weakPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 95% strength, try including more special characters, using numbers or using a longer password"), err)
	}

	err = validate.Password(shortPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 39% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.Password(specialPassword)
	if suite.Error(err) {
		suite.Equal(errors.New("password is only 53% strength, try including more special characters or using a longer password"), err)
	}

	err = validate.Password(longPassword)
	if suite.NoError(err) {
		suite.Equal(nil, err)
	}

	err = validate.Password(tooLong)
	if suite.Error(err) {
		suite.Equal(errors.New("password should be no more than 72 bytes, provided password was 571 bytes"), err)
	}

	err = validate.Password(strongPassword)
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
		suite.True("mail: no angle-addr" == err.Error() ||
			// golang 1.21.8 fixed some inconsistencies in net/mail which leads
			// to different error messages.
			"mail: missing word in phrase: mail: invalid string" == err.Error())
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
	testCases := []struct {
		name, input, expected, err string
	}{
		{name: "empty", err: "no language provided"},
		{name: "notALanguage", input: "this isn't a language at all!", err: "language: tag is not well-formed"},
		{name: "english", input: "en", expected: "en"},
		// Should be all lowercase
		{name: "capitalEnglish", input: "EN", expected: "en"},
		// Overlong, should be in ISO 639-1 format
		{name: "arabic3Letters", input: "ara", expected: "ar"},
		// Should be all lowercase
		{name: "mixedCapsEnglish", input: "eN", expected: "en"},
		// Region should be capitalized
		{name: "englishUS", input: "en-us", expected: "en-US"},
		{name: "dutch", input: "nl", expected: "nl"},
		{name: "german", input: "de", expected: "de"},
		{name: "chinese", input: "zh", expected: "zh"},
		{name: "chineseSimplified", input: "zh-Hans", expected: "zh-Hans"},
		{name: "chineseTraditional", input: "zh-Hant", expected: "zh-Hant"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		suite.Run(testCase.name, func() {
			actual, actualErr := validate.Language(testCase.input)
			if testCase.err == "" {
				suite.Equal(testCase.expected, actual)
				suite.NoError(actualErr)
			} else {
				suite.Empty(actual)
				suite.EqualError(actualErr, testCase.err)
			}
		})
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

func (suite *ValidationTestSuite) TestValidateEmojiShortcode() {
	type testStruct struct {
		shortcode string
		ok        bool
	}

	for _, test := range []testStruct{
		{
			shortcode: "peepee",
			ok:        true,
		},
		{
			shortcode: "poo-poo",
			ok:        false,
		},
		{
			shortcode: "-peepee",
			ok:        false,
		},
		{
			shortcode: "p",
			ok:        true,
		},
		{
			shortcode: "pp",
			ok:        true,
		},
		{
			shortcode: "6969",
			ok:        true,
		},
		{
			shortcode: "__peepee",
			ok:        true,
		},
		{
			shortcode: "_",
			ok:        true,
		},
		{
			shortcode: "",
			ok:        false,
		},
		{
			// Too long.
			shortcode: "_XxX_Ultimate_Gamer_dude_6969_420_",
			ok:        false,
		},
		{
			shortcode: "_XxX_Ultimate_Gamer_dude_6969_",
			ok:        true,
		},
	} {
		err := validate.EmojiShortcode(test.shortcode)
		ok := err == nil
		if !suite.Equal(test.ok, ok) {
			suite.T().Logf("fail on %s", test.shortcode)
		}
	}
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}
