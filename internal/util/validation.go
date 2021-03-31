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
	"regexp"

	pwv "github.com/wagslane/go-password-validator"
	"golang.org/x/text/language"
)

const (
	// MinimumPasswordEntropy dictates password strength. See https://github.com/wagslane/go-password-validator
	MinimumPasswordEntropy = 60
	// MinimumReasonLength is the length of chars we expect as a bare minimum effort
	MinimumReasonLength = 40
	// MaximumReasonLength is the maximum amount of chars we're happy to accept
	MaximumReasonLength = 500
	// MaximumEmailLength is the maximum length of an email address we're happy to accept
	MaximumEmailLength = 256
	// MaximumUsernameLength is the maximum length of a username we're happy to accept
	MaximumUsernameLength = 64
	// MaximumPasswordLength is the maximum length of a password we're happy to accept
	MaximumPasswordLength = 64
	// NewUsernameRegexString is string representation of the regular expression for validating usernames
	NewUsernameRegexString = `^[a-z0-9_]+$`
)

var (
	// NewUsernameRegex is the compiled regex for validating new usernames
	NewUsernameRegex = regexp.MustCompile(NewUsernameRegexString)
)

// ValidateNewPassword returns an error if the given password is not sufficiently strong, or nil if it's ok.
func ValidateNewPassword(password string) error {
	if password == "" {
		return errors.New("no password provided")
	}

	if len(password) > MaximumPasswordLength {
		return fmt.Errorf("password should be no more than %d chars", MaximumPasswordLength)
	}

	return pwv.Validate(password, MinimumPasswordEntropy)
}

// ValidateUsername makes sure that a given username is valid (ie., letters, numbers, underscores, check length).
// Returns an error if not.
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("no username provided")
	}

	if len(username) > MaximumUsernameLength {
		return fmt.Errorf("username should be no more than %d chars but '%s' was %d", MaximumUsernameLength, username, len(username))
	}

	if !NewUsernameRegex.MatchString(username) {
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

	if len(email) > MaximumEmailLength {
		return fmt.Errorf("email address should be no more than %d chars but '%s' was %d", MaximumEmailLength, email, len(email))
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

	if len(reason) < MinimumReasonLength {
		return fmt.Errorf("reason should be at least %d chars but '%s' was %d", MinimumReasonLength, reason, len(reason))
	}

	if len(reason) > MaximumReasonLength {
		return fmt.Errorf("reason should be no more than %d chars but given reason was %d", MaximumReasonLength, len(reason))
	}
	return nil
}

func ValidateDisplayName(displayName string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

func ValidateNote(note string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}
