/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package email

import (
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

// https://datatracker.ietf.org/doc/html/rfc2822
// I did not read the RFC, I just copy and pasted from
// https://pkg.go.dev/net/smtp#SendMail
// and it did seem to work.
func assembleMessage(mailSubject string, mailBody string, mailTo string, mailFrom string) ([]byte, error) {
	if strings.Contains(mailSubject, "\r") || strings.Contains(mailSubject, "\n") {
		return nil, errors.New("email subject must not contain newline characters")
	}

	if strings.Contains(mailFrom, "\r") || strings.Contains(mailFrom, "\n") {
		return nil, errors.New("email from address must not contain newline characters")
	}

	if strings.Contains(mailTo, "\r") || strings.Contains(mailTo, "\n") {
		return nil, errors.New("email to address must not contain newline characters")
	}

	// normalize the message body to use CRLF line endings
	mailBody = strings.ReplaceAll(mailBody, "\r\n", "\n")
	mailBody = strings.ReplaceAll(mailBody, "\n", "\r\n")

	msg := []byte(
		"To: " + mailTo + "\r\n" +
			"Subject: " + mailSubject + "\r\n" +
			"\r\n" +
			mailBody + "\r\n",
	)

	return msg, nil
}
