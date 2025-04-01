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

package gtsmodel

import (
	"net"
	"time"
)

type User struct {
	ID                     string    `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`
	CreatedAt              time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	UpdatedAt              time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`
	Email                  string    `bun:",nullzero,unique"`
	AccountID              string    `bun:"type:CHAR(26),nullzero,notnull,unique"`
	EncryptedPassword      string    `bun:",nullzero,notnull"`
	TwoFactorSecret        string    `bun:",nullzero"`
	TwoFactorBackups       []string  `bun:",nullzero,array"`
	TwoFactorEnabledAt     time.Time `bun:"type:timestamptz,nullzero"`
	SignUpIP               net.IP    `bun:",nullzero"`
	InviteID               string    `bun:"type:CHAR(26),nullzero"`
	Reason                 string    `bun:",nullzero"`
	Locale                 string    `bun:",nullzero"`
	CreatedByApplicationID string    `bun:"type:CHAR(26),nullzero"`
	LastEmailedAt          time.Time `bun:"type:timestamptz,nullzero"`
	ConfirmationToken      string    `bun:",nullzero"`
	ConfirmationSentAt     time.Time `bun:"type:timestamptz,nullzero"`
	ConfirmedAt            time.Time `bun:"type:timestamptz,nullzero"`
	UnconfirmedEmail       string    `bun:",nullzero"`
	Moderator              *bool     `bun:",nullzero,notnull,default:false"`
	Admin                  *bool     `bun:",nullzero,notnull,default:false"`
	Disabled               *bool     `bun:",nullzero,notnull,default:false"`
	Approved               *bool     `bun:",nullzero,notnull,default:false"`
	ResetPasswordToken     string    `bun:",nullzero"`
	ResetPasswordSentAt    time.Time `bun:"type:timestamptz,nullzero"`
	ExternalID             string    `bun:",nullzero,unique"`
}
