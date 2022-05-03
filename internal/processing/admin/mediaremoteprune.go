/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

func (p *processor) MediaRemotePrune(ctx context.Context, mediaRemoteCacheDays int) gtserror.WithCode {
	if mediaRemoteCacheDays < 0 {
		err := fmt.Errorf("invalid value for mediaRemoteCacheDays prune: value was %d, cannot be less than 0", mediaRemoteCacheDays)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	go func() {
		pruned, err := p.mediaManager.PruneRemote(ctx, mediaRemoteCacheDays)
		if err != nil {
			logrus.Errorf("MediaRemotePrune: error pruning: %s", err)
		} else {
			logrus.Infof("MediaRemotePrune: pruned %d entries", pruned)
		}
	}()

	return nil
}
