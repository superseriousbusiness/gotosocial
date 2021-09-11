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

package validate

import (
	"errors"
	"fmt"
	"net/mail"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	pwv "github.com/wagslane/go-password-validator"
	"golang.org/x/text/language"
)

const (
	maximumPasswordLength         = 64
	minimumPasswordEntropy        = 60 // dictates password strength. See https://github.com/wagslane/go-password-validator
	minimumReasonLength           = 40
	maximumReasonLength           = 500
	maximumSiteTitleLength        = 40
	maximumShortDescriptionLength = 500
	maximumDescriptionLength      = 5000
	maximumSiteTermsLength        = 5000
	maximumUsernameLength         = 64
	// maximumEmojiShortcodeLength   = 30
	// maximumHashtagLength          = 30
)

// NewPassword returns an error if the given password is not sufficiently strong, or nil if it's ok.
func NewPassword(password string) error {
	if password == "" {
		return errors.New("no password provided")
	}

	if len(password) > maximumPasswordLength {
		return fmt.Errorf("password should be no more than %d chars", maximumPasswordLength)
	}

	return pwv.Validate(password, minimumPasswordEntropy)
}

// Username makes sure that a given username is valid (ie., letters, numbers, underscores, check length).
// Returns an error if not.
func Username(username string) error {
	if username == "" {
		return errors.New("no username provided")
	}

	if !regexes.Username.MatchString(username) {
		return fmt.Errorf("given username %s was invalid: must contain only lowercase letters, numbers, and underscores, max %d characters", username, maximumUsernameLength)
	}

	return nil
}

// Email makes sure that a given email address is a valid address.
// Returns an error if not.
func Email(email string) error {
	if email == "" {
		return errors.New("no email provided")
	}

	_, err := mail.ParseAddress(email)
	return err
}

// Language checks that the given language string is a 2- or 3-letter ISO 639 code.
// Returns an error if the language cannot be parsed. See: https://pkg.go.dev/golang.org/x/text/language
func Language(lang string) error {
	if lang == "" {
		return errors.New("no language provided")
	}
	_, err := language.ParseBase(lang)
	return err
}

// SignUpReason checks that a sufficient reason is given for a server signup request
func SignUpReason(reason string, reasonRequired bool) error {
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

// DisplayName checks that a requested display name is valid
func DisplayName(displayName string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

// Note checks that a given profile/account note/bio is valid
func Note(note string) error {
	// TODO: add some validation logic here -- length, characters, etc
	return nil
}

// Privacy checks that the desired privacy setting is valid
func Privacy(privacy string) error {
	if privacy == "" {
		return fmt.Errorf("empty string for privacy not allowed")
	}
	switch apimodel.Visibility(privacy) {
	case apimodel.VisibilityDirect, apimodel.VisibilityMutualsOnly, apimodel.VisibilityPrivate, apimodel.VisibilityPublic, apimodel.VisibilityUnlisted:
		return nil
	}
	return fmt.Errorf("privacy %s was not recognized", privacy)
}

// EmojiShortcode just runs the given shortcode through the regular expression
// for emoji shortcodes, to figure out whether it's a valid shortcode, ie., 2-30 characters,
// lowercase a-z, numbers, and underscores.
func EmojiShortcode(shortcode string) error {
	if !regexes.EmojiShortcode.MatchString(shortcode) {
		return fmt.Errorf("shortcode %s did not pass validation, must be between 2 and 30 characters, lowercase letters, numbers, and underscores only", shortcode)
	}
	return nil
}

// SiteTitle ensures that the given site title is within spec.
func SiteTitle(siteTitle string) error {
	if len(siteTitle) > maximumSiteTitleLength {
		return fmt.Errorf("site title should be no more than %d chars but given title was %d", maximumSiteTitleLength, len(siteTitle))
	}

	return nil
}

// SiteShortDescription ensures that the given site short description is within spec.
func SiteShortDescription(d string) error {
	if len(d) > maximumShortDescriptionLength {
		return fmt.Errorf("short description should be no more than %d chars but given description was %d", maximumShortDescriptionLength, len(d))
	}

	return nil
}

// SiteDescription ensures that the given site description is within spec.
func SiteDescription(d string) error {
	if len(d) > maximumDescriptionLength {
		return fmt.Errorf("description should be no more than %d chars but given description was %d", maximumDescriptionLength, len(d))
	}

	return nil
}

// SiteTerms ensures that the given site terms string is within spec.
func SiteTerms(t string) error {
	if len(t) > maximumSiteTermsLength {
		return fmt.Errorf("terms should be no more than %d chars but given terms was %d", maximumSiteTermsLength, len(t))
	}

	return nil
}

// ULID returns true if the passed string is a valid ULID.
func ULID(i string) bool {
	return regexes.ULID.MatchString(i)
}
