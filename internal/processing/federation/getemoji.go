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

package federation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) GetEmoji(ctx context.Context, requestedEmojiID string, requestURL *url.URL) (interface{}, gtserror.WithCode) {
	if _, errWithCode := p.federator.AuthenticateFederatedRequest(ctx, ""); errWithCode != nil {
		return nil, errWithCode
	}

	requestedEmoji, err := p.db.GetEmojiByID(ctx, requestedEmojiID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("database error getting emoji with id %s: %s", requestedEmojiID, err))
	}

	if requestedEmoji.Domain != "" {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji with id %s doesn't belong to this instance (domain %s)", requestedEmojiID, requestedEmoji.Domain))
	}

	if *requestedEmoji.Disabled {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("emoji with id %s has been disabled", requestedEmojiID))
	}

	apEmoji, err := p.tc.EmojiToAS(ctx, requestedEmoji)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting gtsmodel emoji with id %s to ap emoji: %s", requestedEmojiID, err))
	}

	data, err := streams.Serialize(apEmoji)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return data, nil
}
