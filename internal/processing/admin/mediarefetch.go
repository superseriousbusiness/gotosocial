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

package admin

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (p *processor) MediaRefetch(ctx context.Context, requestingAccount *gtsmodel.Account, domain string) gtserror.WithCode {
	transport, err := p.transportController.NewTransportForUsername(ctx, requestingAccount.Username)
	if err != nil {
		err = fmt.Errorf("error getting transport for user %s during media refetch request: %w", requestingAccount.Username, err)
		return gtserror.NewErrorInternalError(err)
	}

	go func() {
		log.Info(ctx, "starting emoji refetch")
		refetched, err := p.mediaManager.RefetchEmojis(context.Background(), domain, transport.DereferenceMedia)
		if err != nil {
			log.Errorf(ctx, "error refetching emojis: %s", err)
		} else {
			log.Infof(ctx, "refetched %d emojis from remote", refetched)
		}
	}()

	return nil
}
