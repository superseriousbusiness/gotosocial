/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) PostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	// pass the fromFederator channel through to postInbox, since it'll be needed later
	contextWithChannel := context.WithValue(ctx, util.APFromFederatorChanKey, p.fromFederator)
	return p.federator.FederatingActor().PostInbox(contextWithChannel, w, r)
}
