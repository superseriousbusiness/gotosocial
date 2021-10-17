/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

// NewNoopSender returns a no-op email sender that will just execute the given sendCallback
// every time it would otherwise send an email.
//
// The 'data' parameter in the callback will be either a ConfirmData or a ResetData struct.
//
// Passing a nil function is also acceptable, in which case the send functions will just return nil.
func NewNoopSender(sendCallback func (toAddress string, data interface{})) Sender {
	return &noopSender{
		sendCallback: sendCallback,
	}
}

type noopSender struct {
	sendCallback func (toAddress string, data interface{})
}

func (s *noopSender) SendConfirmEmail(toAddress string, data ConfirmData) error {
	if s.sendCallback != nil {
		s.sendCallback(toAddress, data)
	}
	return nil
}

func (s *noopSender) SendResetEmail(toAddress string, data ResetData) error {
	if s.sendCallback != nil {
		s.sendCallback(toAddress, data)
	}
	return nil
}

func (s *noopSender) ExecuteTemplate(templateName string, data interface{}) (string, error) {
	return "", nil
}

func (s *noopSender) AssembleMessage(mailSubject string, mailBody string, mailTo string) []byte {
	return []byte{}
}
