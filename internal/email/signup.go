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

var (
	newSignupTemplate = "email_new_signup.tmpl"
	newSignupSubject  = "GoToSocial New Sign-Up"
)

type NewSignupData struct {
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// Email address sign-up was created with.
	SignupEmail string
	// Username submitted on the sign-up form.
	SignupUsername string
	// Reason given on the sign-up form.
	SignupReason string
	// URL to open the sign-up in the settings panel.
	SignupURL string
}

func (s *sender) SendNewSignupEmail(toAddresses []string, data NewSignupData) error {
	return s.sendTemplate(newSignupTemplate, newSignupSubject, data, toAddresses...)
}

var (
	signupApprovedTemplate = "email_signup_approved.tmpl"
	signupApprovedSubject  = "GoToSocial Sign-Up Approved"
)

type SignupApprovedData struct {
	// Username to be addressed.
	Username string
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
}

func (s *sender) SendSignupApprovedEmail(toAddress string, data SignupApprovedData) error {
	return s.sendTemplate(signupApprovedTemplate, signupApprovedSubject, data, toAddress)
}

var (
	signupRejectedTemplate = "email_signup_rejected.tmpl"
	signupRejectedSubject  = "GoToSocial Sign-Up Rejected"
)

type SignupRejectedData struct {
	// Message to the rejected applicant.
	Message string
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
}

func (s *sender) SendSignupRejectedEmail(toAddress string, data SignupRejectedData) error {
	return s.sendTemplate(signupRejectedTemplate, signupRejectedSubject, data, toAddress)
}
