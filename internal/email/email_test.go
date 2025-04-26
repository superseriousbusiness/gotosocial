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

package email_test

import (
	"regexp"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type EmailTestSuite struct {
	suite.Suite

	sender email.Sender

	sentEmails map[string]string
}

func (suite *EmailTestSuite) SetupTest() {
	testrig.InitTestLog()
	suite.sentEmails = make(map[string]string)
	suite.sender = testrig.NewEmailSender("../../web/template/", suite.sentEmails)
}

// strips non deteministic headers from mails
func (suite *EmailTestSuite) stripHeaders() {
	re := regexp.MustCompile(`(?m)^(Date:|Message-ID:) .*$\n`)
	for key, mail := range suite.sentEmails {
		res := re.ReplaceAllString(mail, "")
		suite.sentEmails[key] = res
	}
}

func (suite *EmailTestSuite) TestTemplateConfirmNewSignup() {
	confirmData := email.ConfirmData{
		Username:     "test",
		InstanceURL:  "https://example.org",
		InstanceName: "Test Instance",
		ConfirmLink:  "https://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa",
		NewSignup:    true,
	}

	suite.sender.SendConfirmEmail("user@example.org", confirmData)
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial Email Confirmation\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello test!\r\n\r\nYou are receiving this mail because you've requested an account on https://example.org.\r\n\r\nTo use your account, you must confirm that this is your email address.\r\n\r\nTo confirm your email, paste the following in your browser's address bar:\r\n\r\nhttps://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\r\n\r\n---\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateConfirm() {
	confirmData := email.ConfirmData{
		Username:     "test",
		InstanceURL:  "https://example.org",
		InstanceName: "Test Instance",
		ConfirmLink:  "https://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa",
		NewSignup:    false,
	}

	suite.sender.SendConfirmEmail("user@example.org", confirmData)
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial Email Confirmation\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello test!\r\n\r\nYou are receiving this mail because you've requested an email address change on https://example.org.\r\n\r\nTo complete the change, you must confirm that this is your email address.\r\n\r\nTo confirm your email, paste the following in your browser's address bar:\r\n\r\nhttps://example.org/confirm_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\r\n\r\n---\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReset() {
	resetData := email.ResetData{
		Username:     "test",
		InstanceURL:  "https://example.org",
		InstanceName: "Test Instance",
		ResetLink:    "https://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa",
	}

	suite.sender.SendResetEmail("user@example.org", resetData)
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial Password Reset\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello test!\r\n\r\nYou are receiving this mail because a password reset has been requested for your account on https://example.org.\r\n\r\nTo reset your password, paste the following in your browser's address bar:\r\n\r\nhttps://example.org/reset_email?token=ee24f71d-e615-43f9-afae-385c0799b7fa\r\n\r\n---\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportRemoteToLocal() {
	// Someone from a remote instance has reported one of our users.
	reportData := email.NewReportData{
		InstanceURL:        "https://example.org",
		InstanceName:       "Test Instance",
		ReportURL:          "https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R",
		ReportDomain:       "fossbros-anonymous.io",
		ReportTargetDomain: "",
	}

	if err := suite.sender.SendNewReportEmail([]string{"user@example.org"}, reportData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial New Report\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello moderator of Test Instance (https://example.org)!\r\n\r\nSomeone from fossbros-anonymous.io has reported a user from your instance.\r\n\r\nTo view the report, paste the following link into your browser: https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportLocalToRemote() {
	// Someone from our instance has reported a remote user.
	reportData := email.NewReportData{
		InstanceURL:        "https://example.org",
		InstanceName:       "Test Instance",
		ReportURL:          "https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R",
		ReportDomain:       "",
		ReportTargetDomain: "fossbros-anonymous.io",
	}

	if err := suite.sender.SendNewReportEmail([]string{"user@example.org"}, reportData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial New Report\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello moderator of Test Instance (https://example.org)!\r\n\r\nSomeone from your instance has reported a user from fossbros-anonymous.io.\r\n\r\nTo view the report, paste the following link into your browser: https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportLocalToLocal() {
	// Someone from our instance has reported another user on our instance.
	reportData := email.NewReportData{
		InstanceURL:        "https://example.org",
		InstanceName:       "Test Instance",
		ReportURL:          "https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R",
		ReportDomain:       "",
		ReportTargetDomain: "",
	}

	if err := suite.sender.SendNewReportEmail([]string{"user@example.org"}, reportData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial New Report\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello moderator of Test Instance (https://example.org)!\r\n\r\nSomeone from your instance has reported another user from your instance.\r\n\r\nTo view the report, paste the following link into your browser: https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportMoreThanOneModeratorAddress() {
	reportData := email.NewReportData{
		InstanceURL:        "https://example.org",
		InstanceName:       "Test Instance",
		ReportURL:          "https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R",
		ReportDomain:       "fossbros-anonymous.io",
		ReportTargetDomain: "",
	}

	// Send the email to multiple addresses
	if err := suite.sender.SendNewReportEmail([]string{"user@example.org", "admin@example.org"}, reportData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: Undisclosed Recipients:;\r\nFrom: test@example.org\r\nSubject: GoToSocial New Report\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello moderator of Test Instance (https://example.org)!\r\n\r\nSomeone from fossbros-anonymous.io has reported a user from your instance.\r\n\r\nTo view the report, paste the following link into your browser: https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportMoreThanOneModeratorAddressDisclose() {
	config.SetSMTPDiscloseRecipients(true)

	reportData := email.NewReportData{
		InstanceURL:        "https://example.org",
		InstanceName:       "Test Instance",
		ReportURL:          "https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R",
		ReportDomain:       "fossbros-anonymous.io",
		ReportTargetDomain: "",
	}

	// Send the email to multiple addresses
	if err := suite.sender.SendNewReportEmail([]string{"user@example.org", "admin@example.org"}, reportData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org, admin@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial New Report\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello moderator of Test Instance (https://example.org)!\r\n\r\nSomeone from fossbros-anonymous.io has reported a user from your instance.\r\n\r\nTo view the report, paste the following link into your browser: https://example.org/settings/admin/reports/01GVJHN1RTYZCZTCXVPPPKBX6R\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportClosedOK() {
	reportClosedData := email.ReportClosedData{
		InstanceURL:          "https://example.org",
		InstanceName:         "Test Instance",
		ReportTargetUsername: "foss_satan",
		ReportTargetDomain:   "fossbros-anonymous.io",
		ActionTakenComment:   "User was yeeted. Thank you for reporting!",
	}

	if err := suite.sender.SendReportClosedEmail("user@example.org", reportClosedData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial Report Closed\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello !\r\n\r\nYou recently reported the account @foss_satan@fossbros-anonymous.io to the moderator(s) of Test Instance (https://example.org).\r\n\r\nThe report you submitted has now been closed.\r\n\r\nThe moderator who closed the report left the following comment: User was yeeted. Thank you for reporting!\r\n\r\n---\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func (suite *EmailTestSuite) TestTemplateReportClosedLocalAccountNoComment() {
	reportClosedData := email.ReportClosedData{
		InstanceURL:          "https://example.org",
		InstanceName:         "Test Instance",
		ReportTargetUsername: "1happyturtle",
		ReportTargetDomain:   "",
		ActionTakenComment:   "",
	}

	if err := suite.sender.SendReportClosedEmail("user@example.org", reportClosedData); err != nil {
		suite.FailNow(err.Error())
	}
	suite.stripHeaders()
	suite.Len(suite.sentEmails, 1)
	suite.Equal("To: user@example.org\r\nFrom: test@example.org\r\nSubject: GoToSocial Report Closed\r\nMIME-Version: 1.0\r\nContent-Transfer-Encoding: 8bit\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nHello !\r\n\r\nYou recently reported the account @1happyturtle to the moderator(s) of Test Instance (https://example.org).\r\n\r\nThe report you submitted has now been closed.\r\n\r\nThe moderator who closed the report did not leave a comment.\r\n\r\n---\r\n\r\nIf you believe you've been sent this email in error, feel free to ignore it, or contact the administrator of https://example.org.\r\n\r\n", suite.sentEmails["user@example.org"])
}

func TestEmailTestSuite(t *testing.T) {
	suite.Run(t, new(EmailTestSuite))
}
