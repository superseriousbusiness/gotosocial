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

package user

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"image/png"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-byteutil"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

var b32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

func (p *Processor) TwoFactorQRCodePngGet(
	ctx context.Context,
	user *gtsmodel.User,
) (*apimodel.Content, gtserror.WithCode) {
	// Get the 2FA url for this user.
	totpURI, errWithCode := p.TwoFactorQRCodeURIGet(ctx, user)
	if errWithCode != nil {
		return nil, errWithCode
	}

	key, err := otp.NewKeyFromURL(totpURI.String())
	if err != nil {
		err := gtserror.Newf("error creating totp key from url: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Spawn a QR code image from the key.
	qr, err := key.Image(256, 256)
	if err != nil {
		err := gtserror.Newf("error creating qr image from key: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Blat the key into a buffer.
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, qr); err != nil {
		err := gtserror.Newf("error encoding qr image to png: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Return it as our nice content model.
	return &apimodel.Content{
		ContentType:   "image/png",
		ContentLength: int64(buf.Len()),
		Content:       io.NopCloser(buf),
	}, nil
}

// TwoFactorQRCodeURIGet will generate a new
// 2 factor auth secret for user, and return a
// URI of expected format for generating a QR code
// or inputting into a password manager.
//
// This may be called multiple times without error
// UNTIL the moment the user has finalized enabling
// 2FA. i.e. when user.TwoFactorEnabled() == true.
// Until this point, the URI may be requested for
// both QR code generation, and requesting the URI,
// but once 2FA is confirmed enabled it is not safe
// to re-share the agreed-upon secret.
func (p *Processor) TwoFactorQRCodeURIGet(
	ctx context.Context,
	user *gtsmodel.User,
) (*url.URL, gtserror.WithCode) {
	if user.TwoFactorEnabled() {
		const errText = "2fa already enabled; not sharing secret again"
		return nil, gtserror.NewErrorConflict(errors.New(errText), errText)
	}

	// Only generate new 2FA secret
	// if not already been generated
	// during this enabling process.
	if user.TwoFactorSecret == "" {

		// 32 bytes should be plenty entropy.
		secret := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, secret); err != nil {
			err := gtserror.Newf("error generating new secret: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Set + store the secret.
		user.TwoFactorSecret = b32NoPadding.EncodeToString(secret)
		if err := p.state.DB.UpdateUser(ctx, user, "two_factor_secret"); err != nil {
			err := gtserror.Newf("db error updating user: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// see: https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	issuer := config.GetHost() + " - GoToSocial"
	return &url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   "/" + issuer + ":" + user.Email,
		RawQuery: encodeQuery(url.Values{
			"secret":    {user.TwoFactorSecret},
			"issuer":    {issuer},
			"period":    {"30s"}, // 30s totp validity.
			"digits":    {"6"},   // 6-digit totp.
			"algorithm": {"SHA1"},
		}),
	}, nil
}

// TwoFactorEnable will enable 2 factor auth for
// account, using given TOTP code to validate the
// user's 2fa secret before continuing.
func (p *Processor) TwoFactorEnable(
	ctx context.Context,
	user *gtsmodel.User,
	code string,
) ([]string, gtserror.WithCode) {
	if user.TwoFactorEnabled() {
		const errText = "2fa already enabled; disable it first then try again"
		return nil, gtserror.NewErrorConflict(errors.New(errText), errText)
	}

	if user.TwoFactorSecret == "" {
		const errText = "no 2fa secret stored; first read qr code / totp secret"
		return nil, gtserror.NewErrorForbidden(errors.New(errText), errText)
	}

	// Try validating the provided code and give
	// a helpful error message if it doesn't work.
	if !totp.Validate(code, user.TwoFactorSecret) {
		const errText = "invalid code provided, you may have been too late, try again; " +
			"if it keeps not working, pester your admin to check that the server clock is correct"
		return nil, gtserror.NewErrorForbidden(errors.New(errText), errText)
	}

	// Valid code was provided so we
	// should turn 2fa on for this user.
	user.TwoFactorEnabledAt = time.Now()

	// Create recovery codes in cleartext
	// to show to the user ONCE ONLY.
	backupsClearText := make([]string, 8)
	for i := 0; i < 8; i++ {
		backupsClearText[i] = util.MustGenerateSecret()
	}

	// Store only the bcrypt-encrypted
	// versions of the recovery codes.
	user.TwoFactorBackups = make([]string, 8)
	for i, backup := range backupsClearText {
		encryptedBackup, err := bcrypt.GenerateFromPassword(
			byteutil.S2B(backup),
			bcrypt.DefaultCost,
		)
		if err != nil {
			err := gtserror.Newf("error encrypting backup codes: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		user.TwoFactorBackups[i] = string(encryptedBackup)
	}

	// Update user in the database.
	if err := p.state.DB.UpdateUser(ctx,
		user,
		"two_factor_enabled_at",
		"two_factor_backups",
	); err != nil {
		err := gtserror.Newf("db error updating user: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return backupsClearText, nil
}

// TwoFactorDisable: see TwoFactorDisable().
func (p *Processor) TwoFactorDisable(
	ctx context.Context,
	user *gtsmodel.User,
	password string,
) gtserror.WithCode {
	// Ensure provided password is correct.
	if err := bcrypt.CompareHashAndPassword(
		byteutil.S2B(user.EncryptedPassword),
		byteutil.S2B(password),
	); err != nil {
		const errText = "incorrect password"
		return gtserror.NewErrorUnauthorized(errors.New(errText), errText)
	}

	// Disable 2 factor auth for this account.
	return TwoFactorDisable(ctx, p.state, user)
}

// TwoFactorDisable disables 2 factor auth
// for given user account. Note this should
// be gated with password authentication if
// accessed via web.
func TwoFactorDisable(
	ctx context.Context,
	state *state.State,
	user *gtsmodel.User,
) gtserror.WithCode {
	if !user.TwoFactorEnabled() {
		const errText = "2fa already disabled"
		return gtserror.NewErrorConflict(errors.New(errText), errText)
	}

	// Clear 2FA fields on user account.
	user.TwoFactorEnabledAt = time.Time{}
	user.TwoFactorSecret = ""
	user.TwoFactorBackups = nil
	if err := state.DB.UpdateUser(ctx,
		user,
		"two_factor_enabled_at",
		"two_factor_secret",
		"two_factor_backups",
	); err != nil {
		err := gtserror.Newf("db error updating user: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}

// encodeQuery is a copy-paste of url.Values.Encode, except it uses
// %20 instead of + to encode spaces. This is necessary to correctly
// render spaces in some authenticator apps, like Google Authenticator.
//
// [Note: this func and the above comment are both taken
// directly from github.com/pquerna/otp/internal/encode.go.]
func encodeQuery(v url.Values) string {
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		// Changed from url.QueryEscape.
		keyEscaped := url.PathEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			// Changed from url.QueryEscape.
			buf.WriteString(url.PathEscape(v))
		}
	}
	return buf.String()
}
