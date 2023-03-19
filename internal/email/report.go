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
	newReportTemplate    = "email_new_report.tmpl"
	newReportSubject     = "GoToSocial New Report"
	reportClosedTemplate = "email_report_closed.tmpl"
	reportClosedSubject  = "GoToSocial Report Closed"
)

type NewReportData struct {
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// URL to open the report in the settings panel.
	ReportURL string
	// Domain from which the report originated.
	// Can be empty string for local reports.
	ReportDomain string
	// Domain targeted by the report.
	// Can be empty string for local reports targeting local users.
	ReportTargetDomain string
}

func (s *sender) SendNewReportEmail(toAddresses []string, data NewReportData) error {
	return s.sendTemplate(newReportTemplate, newReportSubject, data, toAddresses...)
}

type ReportClosedData struct {
	// Username to be addressed.
	Username string
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// Username of the report target.
	ReportTargetUsername string
	// Domain of the report target.
	// Can be empty string for local reports targeting local users.
	ReportTargetDomain string
	// Comment left by the admin who closed the report.
	ActionTakenComment string
}

func (s *sender) SendReportClosedEmail(toAddress string, data ReportClosedData) error {
	return s.sendTemplate(reportClosedTemplate, reportClosedSubject, data, toAddress)
}
