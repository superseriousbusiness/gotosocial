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
	"bytes"
	"net/smtp"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

const (
	confirmTemplate = "email_confirm_text.tmpl"
	confirmSubject  = "GoToSocial Email Confirmation"
)

func (s *sender) SendConfirmEmail(toAddress string, data ConfirmData) error {
	buf := &bytes.Buffer{}
	if err := s.template.ExecuteTemplate(buf, confirmTemplate, data); err != nil {
		return err
	}
	confirmBody := buf.String()

	msg, err := assembleMessage(confirmSubject, confirmBody, toAddress, s.from)
	if err != nil {
		return err
	}
	log.Trace(s.hostAddress + "\n" + config.GetSMTPUsername() + ":password" + "\n" + s.from + "\n" + toAddress + "\n\n" + string(msg) + "\n")
	return smtp.SendMail(s.hostAddress, s.auth, s.from, []string{toAddress}, msg)
}

// ConfirmData represents data passed into the confirm email address template.
type ConfirmData struct {
	// Username to be addressed.
	Username string
	// URL of the instance to present to the receiver.
	InstanceURL string
	// Name of the instance to present to the receiver.
	InstanceName string
	// Link to present to the receiver to click on and do the confirmation.
	// Should be a full link with protocol eg., https://example.org/confirm_email?token=some-long-token
	ConfirmLink string
}
