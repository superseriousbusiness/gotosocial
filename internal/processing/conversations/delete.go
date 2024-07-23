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

package conversations

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *Processor) Delete(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	id string,
) gtserror.WithCode {
	// Get the conversation so that we can check its owning account ID.
	conversation, errWithCode := p.getConversationOwnedBy(gtscontext.SetBarebones(ctx), id, requestingAccount)
	if errWithCode != nil {
		return errWithCode
	}

	// Delete the conversation.
	if err := p.state.DB.DeleteConversationByID(ctx, conversation.ID); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
