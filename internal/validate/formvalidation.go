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

package validate

import (
	"errors"
	"fmt"
	"net/mail"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
	pwv "github.com/wagslane/go-password-validator"
	"golang.org/x/text/language"
)

const (
	maximumPasswordLength         = 72 // 72 bytes is the maximum length afforded by bcrypt. See https://pkg.go.dev/golang.org/x/crypto/bcrypt#GenerateFromPassword.
	minimumPasswordEntropy        = 60 // Heuristic for password strength. See https://github.com/wagslane/go-password-validator.
	minimumReasonLength           = 40
	maximumReasonLength           = 500
	maximumSiteTitleLength        = 40
	maximumShortDescriptionLength = 500
	maximumDescriptionLength      = 5000
	maximumSiteTermsLength        = 5000
	maximumUsernameLength         = 64
	maximumEmojiCategoryLength    = 64
	maximumProfileFieldLength     = 255
	maximumListTitleLength        = 200
	maximumFilterKeywordLength    = 40
	maximumFilterTitleLength      = 200
)

// Password returns a helpful error if the given password
// is too short, too long, or not sufficiently strong.
func Password(password string) error {
	// Ensure length is OK first.
	if pwLen := len(password); pwLen == 0 {
		return errors.New("no password provided / provided password was 0 bytes")
	} else if pwLen > maximumPasswordLength {
		return fmt.Errorf(
			"password should be no more than %d bytes, provided password was %d bytes",
			maximumPasswordLength, pwLen,
		)
	}

	if err := pwv.Validate(password, minimumPasswordEntropy); err != nil {
		// Calculate the percentage of our desired entropy this password fulfils.
		entropyPercent := int(100 * pwv.GetEntropy(password) / minimumPasswordEntropy)

		// Replace the first 17 bytes (`insecure password`)
		// of the error string with our own entropy message.
		entropyMsg := fmt.Sprintf("password is only %d%% strength", entropyPercent)
		errMsg := entropyMsg + err.Error()[17:]

		return errors.New(errMsg)
	}

	return nil // password OK
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

// Language checks that the given language string is a valid, if not necessarily canonical, BCP 47 language tag.
// Returns a canonicalized version of the tag if the language can be parsed.
// Returns an error if the language cannot be parsed.
// See: https://pkg.go.dev/golang.org/x/text/language
func Language(lang string) (string, error) {
	if lang == "" {
		return "", errors.New("no language provided")
	}
	parsed, err := language.Parse(lang)
	if err != nil {
		return "", err
	}
	return parsed.String(), err
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

	length := len([]rune(reason))

	if length < minimumReasonLength {
		return fmt.Errorf("reason should be at least %d chars but '%s' was %d", minimumReasonLength, reason, length)
	}

	if length > maximumReasonLength {
		return fmt.Errorf("reason should be no more than %d chars but given reason was %d", maximumReasonLength, length)
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
	return fmt.Errorf("privacy '%s' was not recognized, valid options are 'direct', 'mutuals_only', 'private', 'public', 'unlisted'", privacy)
}

// StatusContentType checks that the desired status format setting is valid.
func StatusContentType(statusContentType string) error {
	if statusContentType == "" {
		return fmt.Errorf("empty string for status format not allowed")
	}
	switch apimodel.StatusContentType(statusContentType) {
	case apimodel.StatusContentTypePlain, apimodel.StatusContentTypeMarkdown:
		return nil
	}
	return fmt.Errorf("status content type '%s' was not recognized, valid options are 'text/plain', 'text/markdown'", statusContentType)
}

func CustomCSS(customCSS string) error {
	if !config.GetAccountsAllowCustomCSS() {
		return errors.New("accounts-allow-custom-css is not enabled for this instance")
	}

	maximumCustomCSSLength := config.GetAccountsCustomCSSLength()
	if length := len([]rune(customCSS)); length > maximumCustomCSSLength {
		return fmt.Errorf("custom_css must be less than %d characters, but submitted custom_css was %d characters", maximumCustomCSSLength, length)
	}

	return nil
}

func InstanceCustomCSS(customCSS string) error {

	maximumCustomCSSLength := config.GetAccountsCustomCSSLength()
	if length := len([]rune(customCSS)); length > maximumCustomCSSLength {
		return fmt.Errorf("custom_css must be less than %d characters, but submitted custom_css was %d characters", maximumCustomCSSLength, length)
	}

	return nil
}

// EmojiShortcode just runs the given shortcode through the regular expression
// for emoji shortcodes, to figure out whether it's a valid shortcode, ie., 1-30 characters,
// a-zA-Z, numbers, and underscores.
func EmojiShortcode(shortcode string) error {
	if !regexes.EmojiValidator.MatchString(shortcode) {
		return fmt.Errorf("shortcode %s did not pass validation, must be between 1 and 30 characters, letters, numbers, and underscores only", shortcode)
	}
	return nil
}

// EmojiCategory validates the length of the given category string.
func EmojiCategory(category string) error {
	if length := len(category); length > maximumEmojiCategoryLength {
		return fmt.Errorf("emoji category %s did not pass validation, must be less than %d characters, but provided value was %d characters", category, maximumEmojiCategoryLength, length)
	}
	return nil
}

// SiteTitle ensures that the given site title is within spec.
func SiteTitle(siteTitle string) error {
	if length := len([]rune(siteTitle)); length > maximumSiteTitleLength {
		return fmt.Errorf("site title should be no more than %d chars but given title was %d", maximumSiteTitleLength, length)
	}

	return nil
}

// SiteShortDescription ensures that the given site short description is within spec.
func SiteShortDescription(d string) error {
	if length := len([]rune(d)); length > maximumShortDescriptionLength {
		return fmt.Errorf("short description should be no more than %d chars but given description was %d", maximumShortDescriptionLength, length)
	}

	return nil
}

// SiteDescription ensures that the given site description is within spec.
func SiteDescription(d string) error {
	if length := len([]rune(d)); length > maximumDescriptionLength {
		return fmt.Errorf("description should be no more than %d chars but given description was %d", maximumDescriptionLength, length)
	}

	return nil
}

// SiteTerms ensures that the given site terms string is within spec.
func SiteTerms(t string) error {
	if length := len([]rune(t)); length > maximumSiteTermsLength {
		return fmt.Errorf("terms should be no more than %d chars but given terms was %d", maximumSiteTermsLength, length)
	}

	return nil
}

// ULID returns an error if the passed string is not a valid ULID.
// The name param is used to form error messages.
func ULID(i string, name string) error {
	if i == "" {
		return fmt.Errorf("%s must be provided", name)
	}
	if !regexes.ULID.MatchString(i) {
		return fmt.Errorf("%s didn't match the expected ULID format for an ID (26 characters from the set 0123456789ABCDEFGHJKMNPQRSTVWXYZ)", name)
	}
	return nil
}

// ProfileFields validates the length of provided fields slice,
// and also iterates through the fields and trims each name + value
// to maximumProfileFieldLength, if they were above.
func ProfileFields(fields []*gtsmodel.Field) error {
	maximumProfileFields := config.GetAccountsMaxProfileFields()
	if len(fields) > maximumProfileFields {
		return fmt.Errorf("cannot have more than %d profile fields", maximumProfileFields)
	}

	// Trim each field name + value to maximum allowed length.
	for _, field := range fields {
		n := []rune(field.Name)
		if len(n) > maximumProfileFieldLength {
			field.Name = string(n[:maximumProfileFieldLength])
		}

		v := []rune(field.Value)
		if len(v) > maximumProfileFieldLength {
			field.Value = string(v[:maximumProfileFieldLength])
		}
	}

	return nil
}

// ListTitle validates the title of a new or updated List.
func ListTitle(title string) error {
	if title == "" {
		return fmt.Errorf("list title must be provided, and must be no more than %d chars", maximumListTitleLength)
	}

	if length := len([]rune(title)); length > maximumListTitleLength {
		return fmt.Errorf("list title length must be no more than %d chars, provided title was %d chars", maximumListTitleLength, length)
	}

	return nil
}

// ListRepliesPolicy validates the replies_policy of a new or updated list.
func ListRepliesPolicy(repliesPolicy gtsmodel.RepliesPolicy) error {
	switch repliesPolicy {
	case "", gtsmodel.RepliesPolicyFollowed, gtsmodel.RepliesPolicyList, gtsmodel.RepliesPolicyNone:
		// No problem.
		return nil
	default:
		// Uh oh.
		return fmt.Errorf("list replies_policy must be either empty or one of 'followed', 'list', 'none'")
	}
}

// MarkerName checks that the desired marker timeline name is valid.
func MarkerName(name string) error {
	if name == "" {
		return fmt.Errorf("empty string for marker timeline name not allowed")
	}
	switch apimodel.MarkerName(name) {
	case apimodel.MarkerNameHome, apimodel.MarkerNameNotifications:
		return nil
	}
	return fmt.Errorf("marker timeline name '%s' was not recognized, valid options are '%s', '%s'", name, apimodel.MarkerNameHome, apimodel.MarkerNameNotifications)
}

// FilterKeyword validates a filter keyword.
func FilterKeyword(keyword string) error {
	if keyword == "" {
		return fmt.Errorf("filter keyword must be provided, and must be no more than %d chars", maximumFilterKeywordLength)
	}

	if length := len([]rune(keyword)); length > maximumFilterKeywordLength {
		return fmt.Errorf("filter keyword length must be no more than %d chars, provided keyword was %d chars", maximumFilterKeywordLength, length)
	}

	return nil
}

// FilterTitle validates the title of a new or updated filter.
func FilterTitle(title string) error {
	if title == "" {
		return fmt.Errorf("filter title must be provided, and must be no more than %d chars", maximumFilterTitleLength)
	}

	if length := len([]rune(title)); length > maximumFilterTitleLength {
		return fmt.Errorf("filter title length must be no more than %d chars, provided title was %d chars", maximumFilterTitleLength, length)
	}

	return nil
}

// FilterContexts validates the context of a new or updated filter.
func FilterContexts(contexts []apimodel.FilterContext) error {
	if len(contexts) == 0 {
		return fmt.Errorf("at least one filter context is required")
	}
	for _, context := range contexts {
		switch context {
		case apimodel.FilterContextHome,
			apimodel.FilterContextNotifications,
			apimodel.FilterContextPublic,
			apimodel.FilterContextThread,
			apimodel.FilterContextAccount:
			continue
		default:
			return fmt.Errorf(
				"filter context '%s' was not recognized, valid options are '%s', '%s', '%s', '%s', '%s'",
				context,
				apimodel.FilterContextHome,
				apimodel.FilterContextNotifications,
				apimodel.FilterContextPublic,
				apimodel.FilterContextThread,
				apimodel.FilterContextAccount,
			)
		}
	}
	return nil
}

func FilterAction(action apimodel.FilterAction) error {
	switch action {
	case apimodel.FilterActionWarn,
		apimodel.FilterActionHide:
		return nil
	}
	return fmt.Errorf(
		"filter action '%s' was not recognized, valid options are '%s', '%s'",
		action,
		apimodel.FilterActionWarn,
		apimodel.FilterActionHide,
	)
}

// CreateAccount checks through all the prerequisites for
// creating a new account, according to the provided form.
// If the account isn't eligible, an error will be returned.
//
// Side effect: normalizes the provided language tag for the user's locale.
func CreateAccount(form *apimodel.AccountCreateRequest) error {
	if form == nil {
		return errors.New("form was nil")
	}

	if !config.GetAccountsRegistrationOpen() {
		return errors.New("registration is not open for this server")
	}

	if err := Username(form.Username); err != nil {
		return err
	}

	if err := Email(form.Email); err != nil {
		return err
	}

	if err := Password(form.Password); err != nil {
		return err
	}

	if !form.Agreement {
		return errors.New("agreement to terms and conditions not given")
	}

	locale, err := Language(form.Locale)
	if err != nil {
		return err
	}
	form.Locale = locale

	return SignUpReason(form.Reason, config.GetAccountsReasonRequired())
}
