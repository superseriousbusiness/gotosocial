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

package user

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

// DeleteSelf is like Account.Delete, but specifically
// for local user+accounts deleting themselves.
//
// Calling DeleteSelf results in a delete message being enqueued in the processor,
// which causes side effects to occur: delete will be federated out to other instances,
// and the above Delete function will be called afterwards from the processor, to clear
// out the account's bits and bobs, and stubbify it.
func (p *Processor) DeleteSelf(ctx context.Context, account *gtsmodel.Account) gtserror.WithCode {
	// Process the delete side effects asynchronously.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		// Use ap.ObjectProfile here to
		// distinguish this message (user model)
		// from ap.ActorPerson (account model).
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityDelete,
		Origin:         account,
		Target:         account,
	})
	return nil
}
