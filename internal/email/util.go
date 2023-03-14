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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func loadTemplates(templateBaseDir string) (*template.Template, error) {
	if !filepath.IsAbs(templateBaseDir) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current working directory: %s", err)
		}
		templateBaseDir = filepath.Join(cwd, templateBaseDir)
	}

	// look for all templates that start with 'email_'
	return template.ParseGlob(filepath.Join(templateBaseDir, "email_*"))
}

// assembleMessage assembles a valid email message following:
//   - https://datatracker.ietf.org/doc/html/rfc2822
//   - https://pkg.go.dev/net/smtp#SendMail
func assembleMessage(mailSubject string, mailBody string, mailFrom string, mailTo ...string) ([]byte, error) {
	if strings.Contains(mailSubject, "\r") || strings.Contains(mailSubject, "\n") {
		return nil, errors.New("email subject must not contain newline characters")
	}

	if strings.Contains(mailFrom, "\r") || strings.Contains(mailFrom, "\n") {
		return nil, errors.New("email from address must not contain newline characters")
	}

	for _, to := range mailTo {
		if strings.Contains(to, "\r") || strings.Contains(to, "\n") {
			return nil, errors.New("email to address must not contain newline characters")
		}
	}

	// Normalize the message body to use CRLF line endings
	const CRLF = "\r\n"
	mailBody = strings.ReplaceAll(mailBody, CRLF, "\n")
	mailBody = strings.ReplaceAll(mailBody, "\n", CRLF)

	msg := bytes.Buffer{}
	if len(mailTo) == 1 {
		// Address email directly to the one recipient.
		msg.WriteString("To: " + mailTo[0] + CRLF)
	} else {
		// To group, Bcc the multiple recipients.
		msg.WriteString("To: Multiple recipients:;" + CRLF)
		msg.WriteString("Bcc: " + strings.Join(mailTo, ", ") + CRLF)
	}
	msg.WriteString("Subject: " + mailSubject + CRLF)
	msg.WriteString(CRLF)
	msg.WriteString(mailBody)
	msg.WriteString(CRLF)

	return msg.Bytes(), nil
}
