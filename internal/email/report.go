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
	"net/smtp"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

const (
	reportTemplate = "email_report_text.tmpl"
	reportSubject  = "GoToSocial New Moderation Report"
)

type NewReportData struct {
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// URL to open the report in the settings panel.
	ReportURL string
	// Domain from which the report originated.
	ReportDomain string
}

func (s *sender) SendNewReportEmail(toAddresses []string, data NewReportData) error {
	buf := &bytes.Buffer{}
	if err := s.template.ExecuteTemplate(buf, reportTemplate, data); err != nil {
		return err
	}
	reportBody := buf.String()

	msg, err := assembleMessage(reportSubject, reportBody, s.from, toAddresses...)
	if err != nil {
		return err
	}

	if err := smtp.SendMail(s.hostAddress, s.auth, s.from, toAddresses, msg); err != nil {
		return gtserror.SetType(err, gtserror.TypeSMTP)
	}

	return nil
}
