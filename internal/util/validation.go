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

package util

import (
	"errors"
	"fmt"
	"net/mail"

	pwv "github.com/wagslane/go-password-validator"
	"golang.org/x/text/language"
)

// ValidateNewPassword returns an error if the given password is not sufficiently strong, or nil if it's ok.
func ValidateNewPassword(password string) error {
	if password == "" {
		return errors.New("no password provided")
	}

	if len(password) > maximumPasswordLength {
		return fmt.Errorf("password should be no more than %d chars", maximumPasswordLength)
	}

	return pwv.Validate(password, minimumPasswordEntropy)
}

// ValidateUsername makes sure that a given username is valid (ie., letters, numbers, underscores, check length).
// Returns an error if not.
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("no username provided")
	}

	if len(username) > maximumUsernameLength {
		return fmt.Errorf("username should be no more than %d chars but '%s' was %d", maximumUsernameLength, username, len(username))
	}

	if !usernameValidationRegex.MatchString(username) {
		return fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores", username)
	}

	return nil
}

// ValidateEmail makes sure that a given email address is a valid address.
// Returns an error if not.
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("no email provided")
	}

	if len(email) > maximumEmailLength {
		return fmt.Errorf("email address should be no more than %d chars but '%s' was %d", maximumEmailLength, email, len(email))
	}

	_, err := mail.ParseAddress(email)
	return err
}

// ValidateLanguage checks that the given language string is a 2- or 3-letter ISO 639 code.
// Returns an error if the language cannot be parsed. See: https://pkg.go.dev/golang.org/x/text/language
func ValidateLanguage(lang string) error {
	if lang == "" {
		return errors.New("no language provided")
	}
	_, err := language.ParseBase(lang)
	return err
}

// ValidateSignUpReason checks that a sufficient reason is given for a server signup request
func ValidateSignUpReason(reason string, reasonRequired bool) error {
	if !reasonRequired {
		// we don't care!
		// we're not going to do anything with this text anyway if no reason is required
		return nil
	}

	if reason == "" {
		return errors.New("no reason provided")
	}

	if len(reason) < minimumReasonLength {
		return fmt.Errorf("reason should be at least %d chars but '%s' was %d", minimumReasonLength, reason, len(reason))
	}

	if len(reason) > maximumReasonLength {
		return fmt.Errorf("reason should be no more than %d chars but given reason was %d", maximumReasonLength, len(reason))
	}
	return nil
}

// ValidateDisplayName checks that a requested display name is valid
func ValidateDisplayName(displayName string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

// ValidateNote checks that a given profile/account note/bio is valid
func ValidateNote(note string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

// ValidatePrivacy checks that the desired privacy setting is valid
func ValidatePrivacy(privacy string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

// ValidateEmojiShortcode just runs the given shortcode through the regular expression
// for emoji shortcodes, to figure out whether it's a valid shortcode, ie., 2-30 characters,
// lowercase a-z, numbers, and underscores.
func ValidateEmojiShortcode(shortcode string) error {
	if !emojiShortcodeValidationRegex.MatchString(shortcode) {
		return fmt.Errorf("shortcode %s did not pass validation, must be between 2 and 30 characters, lowercase letters, numbers, and underscores only", shortcode)
	}
	return nil
}
