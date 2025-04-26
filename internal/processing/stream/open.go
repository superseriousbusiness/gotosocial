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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/stream"
	"codeberg.org/gruf/go-kv"
)

// Open returns a new Stream for the given account, which will contain a channel for passing messages back to the caller.
func (p *Processor) Open(ctx context.Context, account *gtsmodel.Account, streamType string) (*stream.Stream, gtserror.WithCode) {
	l := log.WithContext(ctx).WithFields(kv.Fields{
		{"account", account.ID},
		{"streamType", streamType},
	}...)
	l.Debug("received open stream request")
	return p.streams.Open(account.ID, streamType), nil
}
