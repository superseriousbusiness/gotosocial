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

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// MediaRefetch forces a refetch of remote emojis.
func (p *Processor) MediaRefetch(ctx context.Context, requestingAccount *gtsmodel.Account, domain string) gtserror.WithCode {
	transport, err := p.transport.NewTransportForUsername(ctx, requestingAccount.Username)
	if err != nil {
		err = fmt.Errorf("error getting transport for user %s during media refetch request: %w", requestingAccount.Username, err)
		return gtserror.NewErrorInternalError(err)
	}

	go func() {
		ctx := gtscontext.WithValues(context.Background(), ctx)
		log.Info(ctx, "starting emoji refetch")
		refetched, err := p.media.RefetchEmojis(ctx, domain, transport.DereferenceMedia)
		if err != nil {
			log.Errorf(ctx, "error refetching emojis: %s", err)
		} else {
			log.Infof(ctx, "refetched %d emojis from remote", refetched)
		}
	}()

	return nil
}

// MediaPrune triggers a non-blocking prune of unused media, orphaned, uncaching remote and fixing cache states.
func (p *Processor) MediaPrune(ctx context.Context, mediaRemoteCacheDays int) gtserror.WithCode {
	if mediaRemoteCacheDays < 0 {
		err := fmt.Errorf("MediaPrune: invalid value for mediaRemoteCacheDays prune: value was %d, cannot be less than 0", mediaRemoteCacheDays)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Start background task performing all media cleanup tasks.
	go func() {
		ctx := context.Background()
		p.cleaner.Media().All(ctx, mediaRemoteCacheDays)
		p.cleaner.Emoji().All(ctx, mediaRemoteCacheDays)
	}()

	return nil
}
