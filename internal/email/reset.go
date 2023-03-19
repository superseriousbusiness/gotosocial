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

package email

const (
	resetTemplate = "email_reset.tmpl"
	resetSubject  = "GoToSocial Password Reset"
)

// ResetData represents data passed into the reset email address template.
type ResetData struct {
	// Username to be addressed.
	Username string
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// Link to present to the receiver to click on and begin the reset process.
	// Should be a full link with protocol eg., https://example.org/reset_password?token=some-reset-password-token
	ResetLink string
}

func (s *sender) SendResetEmail(toAddress string, data ResetData) error {
	return s.sendTemplate(resetTemplate, resetSubject, data, toAddress)
}
