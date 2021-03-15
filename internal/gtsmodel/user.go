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

package gtsmodel

import (
	"net"
	"time"
)

type User struct {
	ID                     string    `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	Email                  string    `pg:",notnull"`
	CreatedAt              time.Time `pg:"type:timestamp,notnull"`
	UpdatedAt              time.Time `pg:"type:timestamp,notnull"`
	EncryptedPassword      string    `pg:",notnull"`
	ResetPasswordToken     string
	ResetPasswordSentAt    time.Time `pg:"type:timestamp"`
	SignInCount            int
	CurrentSignInAt        time.Time `pg:"type:timestamp"`
	LastSignInAt           time.Time `pg:"type:timestamp"`
	CurrentSignInIP        net.IP
	LastSignInIP           net.IP
	Admin                  bool
	ConfirmationToken      string
	ConfirmedAt            time.Time `pg:"type:timestamp"`
	ConfirmationSentAt     time.Time `pg:"type:timestamp"`
	UnconfirmedEmail       string
	Locale                 string
	EncryptedOTPSecret     string
	EncryptedOTPSecretIv   string
	EncryptedOTPSecretSalt string
	ConsumedTimestamp      int
	OTPRequiredForLogin    bool
	LastEmailedAt          time.Time `pg:"type:timestamp"`
	OTPBackupCodes         []string
	FilteredLanguages      []string
	AccountID              string `pg:",notnull"`
	Disabled               bool
	Moderator              bool
	InviteID               string
	RememberToken          string
	ChosenLanguages        []string
	CreatedByApplicationID string
	Approved               bool
	SignInToken            string
	SignInTokenSentAt      time.Time `pg:"type:timestamp"`
	WebauthnID             string
	SignUpIP               net.IP
}
