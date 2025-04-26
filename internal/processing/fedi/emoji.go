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

package fedi

import (
	"context"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
)

// EmojiGet handles the GET for a federated emoji originating from this instance.
func (p *Processor) EmojiGet(ctx context.Context, requestedEmojiID string) (interface{}, gtserror.WithCode) {
	if _, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, ""); errWithCode != nil {
		return nil, errWithCode
	}

	requestedEmoji, err := p.state.DB.GetEmojiByID(ctx, requestedEmojiID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting emoji with id %s: %s", requestedEmojiID, err))
	}

	if !requestedEmoji.IsLocal() {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji with id %s doesn't belong to this instance (domain %s)", requestedEmojiID, requestedEmoji.Domain))
	}

	if *requestedEmoji.Disabled {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji with id %s has been disabled", requestedEmojiID))
	}

	apEmoji, err := p.converter.EmojiToAS(ctx, requestedEmoji)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting gtsmodel emoji with id %s to ap emoji: %s", requestedEmojiID, err))
	}

	data, err := ap.Serialize(apEmoji)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
