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

package admin

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// EmailTest sends a generic test email to the given
// toAddress (which should be a valid email address).
// Message is optional and can be an empty string.
//
// To help callers differentiate between proper errors
// and the smtp errors they're likely fishing for, will
// return 422 + help text on an SMTP error, or 500 otherwise.
func (p *Processor) EmailTest(
	ctx context.Context,
	account *gtsmodel.Account,
	toAddress string,
	message string,
) gtserror.WithCode {
	// Pull our instance entry from the database,
	// so we can greet the email recipient nicely.
	instance, err := p.state.DB.GetInstance(ctx, config.GetHost())
	if err != nil {
		err = fmt.Errorf("SendConfirmEmail: error getting instance: %s", err)
		return gtserror.NewErrorInternalError(err)
	}

	testData := email.TestData{
		SendingUsername: account.Username,
		Message:         message,
		InstanceURL:     instance.URI,
		InstanceName:    instance.Title,
	}

	if err := p.email.SendTestEmail(toAddress, testData); err != nil {
		if gtserror.IsSMTP(err) {
			// An error occurred during the SMTP part.
			// We should indicate this to the caller, as
			// it will likely help them debug the issue.
			return gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}
		// An actual error has occurred.
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
