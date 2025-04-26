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

package stream

import (
	"context"
	"encoding/json"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
	"codeberg.org/gruf/go-byteutil"
)

// StatusUpdate streams the given edited status to any open, appropriate streams belonging to the given account.
func (p *Processor) StatusUpdate(ctx context.Context, account *gtsmodel.Account, status *apimodel.Status, streamType string) {
	b, err := json.Marshal(status)
	if err != nil {
		log.Errorf(ctx, "error marshaling json: %v", err)
		return
	}
	p.streams.Post(ctx, account.ID, stream.Message{
		Payload: byteutil.B2S(b),
		Event:   stream.EventTypeStatusUpdate,
		Stream:  []string{streamType},
	})
}
