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
)

func (p *processor) MediaPrune(ctx context.Context, mediaRemoteCacheDays int) gtserror.WithCode {
	if mediaRemoteCacheDays < 0 {
		err := fmt.Errorf("MediaPrune: invalid value for mediaRemoteCacheDays prune: value was %d, cannot be less than 0", mediaRemoteCacheDays)
		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	if err := p.mediaManager.PruneAll(ctx, mediaRemoteCacheDays, false); err != nil {
		err = fmt.Errorf("MediaPrune: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
