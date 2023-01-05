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

package processing

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
)

func (p *processor) AuthorizeStreamingRequest(ctx context.Context, accessToken string) (*gtsmodel.Account, gtserror.WithCode) {
	return p.streamingProcessor.AuthorizeStreamingRequest(ctx, accessToken)
}

func (p *processor) OpenStreamForAccount(ctx context.Context, account *gtsmodel.Account, streamType string) (*stream.Stream, gtserror.WithCode) {
	return p.streamingProcessor.OpenStreamForAccount(ctx, account, streamType)
}
