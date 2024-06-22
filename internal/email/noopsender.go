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

import (
	"bytes"
	"text/template"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// NewNoopSender returns a no-op email sender that will just execute the given sendCallback
// every time it would otherwise send an email to the given toAddress with the given message value.
//
// Passing a nil function is also acceptable, in which case the send functions will just return nil.
func NewNoopSender(sendCallback func(toAddress string, message string)) (Sender, error) {
	templateBaseDir := config.GetWebTemplateBaseDir()
	msgIDHost := config.GetHost()

	t, err := loadTemplates(templateBaseDir)
	if err != nil {
		return nil, err
	}

	return &noopSender{
		sendCallback: sendCallback,
		msgIDHost:    msgIDHost,
		template:     t,
	}, nil
}

type noopSender struct {
	sendCallback func(toAddress string, message string)
	msgIDHost    string
	template     *template.Template
}

func (s *noopSender) SendConfirmEmail(toAddress string, data ConfirmData) error {
	return s.sendTemplate(confirmTemplate, confirmSubject, data, toAddress)
}

func (s *noopSender) SendResetEmail(toAddress string, data ResetData) error {
	return s.sendTemplate(resetTemplate, resetSubject, data, toAddress)
}

func (s *noopSender) SendTestEmail(toAddress string, data TestData) error {
	return s.sendTemplate(testTemplate, testSubject, data, toAddress)
}

func (s *noopSender) SendNewReportEmail(toAddresses []string, data NewReportData) error {
	return s.sendTemplate(newReportTemplate, newReportSubject, data, toAddresses...)
}

func (s *noopSender) SendReportClosedEmail(toAddress string, data ReportClosedData) error {
	return s.sendTemplate(reportClosedTemplate, reportClosedSubject, data, toAddress)
}

func (s *noopSender) SendNewSignupEmail(toAddresses []string, data NewSignupData) error {
	return s.sendTemplate(newSignupTemplate, newSignupSubject, data, toAddresses...)
}

func (s *noopSender) SendSignupApprovedEmail(toAddress string, data SignupApprovedData) error {
	return s.sendTemplate(signupApprovedTemplate, signupApprovedSubject, data, toAddress)
}

func (s *noopSender) SendSignupRejectedEmail(toAddress string, data SignupRejectedData) error {
	return s.sendTemplate(signupRejectedTemplate, signupRejectedSubject, data, toAddress)
}

func (s *noopSender) sendTemplate(template string, subject string, data any, toAddresses ...string) error {
	buf := &bytes.Buffer{}
	if err := s.template.ExecuteTemplate(buf, template, data); err != nil {
		return err
	}

	msg, err := assembleMessage(subject, buf.String(), "test@example.org", s.msgIDHost, toAddresses...)
	if err != nil {
		return err
	}

	log.Tracef(nil, "NOT SENDING email to %s with contents: %s", toAddresses, msg)

	if s.sendCallback != nil {
		s.sendCallback(toAddresses[0], string(msg))
	}

	return nil
}
