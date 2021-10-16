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

import (
	"bytes"
	"fmt"
)

const (
	mime = `MIME-version: 1.0;
Content-Type: text/html;`
)

func (s *sender) ExecuteTemplate(templateName string, data interface{}) (string, error) {
	buf := &bytes.Buffer{}
	if err := s.template.ExecuteTemplate(buf, templateName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *sender) AssembleMessage(mailSubject string, mailBody string, mailTo string) []byte {
	from := fmt.Sprintf("From: GoToSocial <%s>", s.from)
	to := fmt.Sprintf("To: %s", mailTo)

	msg := []byte(
		mailSubject + "\r\n" +
			from + "\r\n" +
			to + "\r\n" +
			mime + "\r\n" +
			mailBody + "\r\n")

	return msg
}
