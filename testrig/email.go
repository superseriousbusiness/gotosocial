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

package testrig

import (
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// NewEmailSender returns a noop email sender that won't make any remote calls.
//
// If sentEmails is not nil, the noop callback function will place sent emails in
// the map, with email address of the recipient as the key, and the value as the
// parsed email message as it would have been sent.
func NewEmailSender(templateBaseDir string, sentEmails map[string]string) email.Sender {
	config.Config(func(cfg *config.Configuration) {
		cfg.WebTemplateBaseDir = templateBaseDir
	})

	var sendCallback func(toAddress string, message string)

	if sentEmails != nil {
		sendCallback = func(toAddress string, message string) {
			sentEmails[toAddress] = message
		}
	} else {
		sendCallback = func(toAddress string, message string) {
			log.Infof(nil, "Sent email to %s: %s", toAddress, message)
		}
	}

	s, err := email.NewNoopSender(sendCallback)
	if err != nil {
		panic(err)
	}
	return s
}
