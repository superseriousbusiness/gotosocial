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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (s *sender) sendTemplate(template string, subject string, data any, toAddresses ...string) error {
	buf := &bytes.Buffer{}
	if err := s.template.ExecuteTemplate(buf, template, data); err != nil {
		return err
	}

	msg, err := assembleMessage(subject, buf.String(), s.from, s.msgIDHost, toAddresses...)
	if err != nil {
		return err
	}

	if err := SendMail(s.hostname, s.port, s.encryptionMode, s.auth, s.from, toAddresses, msg); err != nil {
		return gtserror.SetSMTP(err)
	}

	return nil
}

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
func assembleMessage(mailSubject string, mailBody string, mailFrom string, msgIDHost string, mailTo ...string) ([]byte, error) {
	if strings.ContainsAny(mailSubject, "\r\n") {
		return nil, errors.New("email subject must not contain newline characters")
	}

	if strings.ContainsAny(mailFrom, "\r\n") {
		return nil, errors.New("email from address must not contain newline characters")
	}

	for _, to := range mailTo {
		if strings.ContainsAny(to, "\r\n") {
			return nil, errors.New("email to address must not contain newline characters")
		}
	}

	// Normalize the message body to use CRLF line endings
	const CRLF = "\r\n"
	mailBody = strings.ReplaceAll(mailBody, CRLF, "\n")
	mailBody = strings.ReplaceAll(mailBody, "\n", CRLF)

	msg := bytes.Buffer{}
	switch {
	case len(mailTo) == 1:
		// Address email directly to the one recipient.
		msg.WriteString("To: " + mailTo[0] + CRLF)
	case config.GetSMTPDiscloseRecipients():
		// Simply address To all recipients.
		msg.WriteString("To: " + strings.Join(mailTo, ", ") + CRLF)
	default:
		// Address To anonymous group.
		//
		// Email will be sent to all recipients but we shouldn't include Bcc header.
		//
		// From the smtp.SendMail function: 'Sending "Bcc" messages is accomplished by
		// including an email address in the to parameter but not including it in the
		// msg headers.'
		msg.WriteString("To: Undisclosed Recipients:;" + CRLF)
	}
	msg.WriteString("Date: " + time.Now().Format(time.RFC822Z) + CRLF)
	msg.WriteString("From: " + mailFrom + CRLF)
	msg.WriteString("Message-ID: <" + uuid.New().String() + "@" + msgIDHost + ">" + CRLF)
	msg.WriteString("Subject: " + mailSubject + CRLF)
	msg.WriteString("MIME-Version: 1.0" + CRLF)
	msg.WriteString("Content-Transfer-Encoding: 8bit" + CRLF)
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"" + CRLF)
	msg.WriteString(CRLF)
	msg.WriteString(mailBody)
	msg.WriteString(CRLF)

	return msg.Bytes(), nil
}

// Slightly modified version of SendMail() from `net/smtp` standard library, in order to support encryptionMode STARTTLS vs. SMTPS
func SendMail(hostname string, port int, encryptionMode string, a smtp.Auth, from string, to []string, msg []byte) error {
	var err error
	if err = validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err = validateLine(recp); err != nil {
			return err
		}
	}
	var c *smtp.Client
	if encryptionMode == "STARTTLS" {
		c, err = Dial(hostname, port)
	} else if encryptionMode == "SMTPS" {
		c, err = DialTLS(hostname, port)
	} else {
		return errors.New(fmt.Sprintf("Unsupported SMTPEncryption '%s'. Only 'SMTPS' and 'STARTTLS' are supported.", encryptionMode))
	}

	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	if encryptionMode == "STARTTLS" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: hostname}
			if err = c.StartTLS(config); err != nil {
				return err
			}
		} else {
			return errors.New(fmt.Sprintf(
				"SMTPEncryption 'STARTTLS' but server '%s:%d' doesn't support the STARTTLS SMTP extension",
				hostname, port,
			))
		}
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// from `net/smtp` standard library
func Dial(hostname string, port int) (*smtp.Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, hostname)
}

// Added to support SMTPS in addition  to STARTTLS
func DialTLS(hostname string, port int) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port), &tls.Config{ServerName: hostname})
	if err != nil {
		return nil, err
	}
	return smtp.NewClient(conn, hostname)
}

// from `net/smtp` standard library
// validateLine checks to see if a line has CR or LF as per RFC 5321.
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
