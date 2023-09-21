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

package federatingdb

import (
	"context"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Announce(ctx context.Context, announce vocab.ActivityStreamsAnnounce) error {
	if log.Level() >= level.DEBUG {
		i, err := marshalItem(announce)
		if err != nil {
			return err
		}
		l := log.WithContext(ctx).
			WithField("announce", i)
		l.Debug("entering Announce")
	}

	receivingAccount, _, internal := extractFromCtx(ctx)
	if internal {
		return nil // Already processed.
	}

	boost, isNew, err := f.converter.ASAnnounceToStatus(ctx, announce)
	if err != nil {
		return gtserror.Newf("error converting announce to boost: %w", err)
	}

	if !isNew {
		// We've already seen this boost;
		// nothing else to do here.
		return nil
	}

	// This is a new boost. Process side effects asynchronously.
	f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
		APObjectType:     ap.ActivityAnnounce,
		APActivityType:   ap.ActivityCreate,
		GTSModel:         boost,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
