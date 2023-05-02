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
	"errors"
	"fmt"

	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// Update sets an existing entry to the database based on the value's
// id.
//
// Note that Activity values received from federated peers may also be
// updated in the database this way if the Federating Protocol is
// enabled. The client may freely decide to store only the id instead of
// the entire value.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Update(ctx context.Context, asType vocab.Type) error {
	l := log.Entry{}.WithContext(ctx)

	if log.Level() >= level.DEBUG {
		i, err := marshalItem(asType)
		if err != nil {
			return err
		}
		l = l.WithField("update", i)
		l.Debug("entering Update")
	}

	receivingAccount, _ := extractFromCtx(ctx)
	if receivingAccount == nil {
		// If the receiving account wasn't set on the context, that means
		// this request didn't pass through the API, but came from inside
		// GtS as the result of another activity on this instance. As such,
		// we must have already processed it in order to reach this stage.
		return nil
	}

	requestingAcctI := ctx.Value(ap.ContextRequestingAccount)
	if requestingAcctI == nil {
		return errors.New("Update: requesting account wasn't set on context")
	}

	requestingAcct, ok := requestingAcctI.(*gtsmodel.Account)
	if !ok {
		return errors.New("Update: requesting account was set on context but couldn't be parsed")
	}

	switch asType.GetTypeName() {
	case ap.ActorApplication, ap.ActorGroup, ap.ActorOrganization, ap.ActorPerson, ap.ActorService:
		return f.updateAccountable(ctx, receivingAccount, requestingAcct, asType)
	}

	return nil
}

func (f *federatingDB) updateAccountable(ctx context.Context, receivingAcct *gtsmodel.Account, requestingAcct *gtsmodel.Account, asType vocab.Type) error {
	accountable, ok := asType.(ap.Accountable)
	if !ok {
		return errors.New("updateAccountable: could not convert vocab.Type to Accountable")
	}

	updatedAcct, err := f.typeConverter.ASRepresentationToAccount(ctx, accountable, "")
	if err != nil {
		return fmt.Errorf("updateAccountable: error converting to account: %w", err)
	}

	if updatedAcct.Domain == config.GetHost() || updatedAcct.Domain == config.GetAccountDomain() {
		// No need to update local accounts; in fact, if we try
		// this it will break the shit out of things so do NOT.
		return nil
	}

	if requestingAcct.URI != updatedAcct.URI {
		return fmt.Errorf("updateAccountable: update for account %s was requested by account %s, this is not valid", updatedAcct.URI, requestingAcct.URI)
	}

	// Set some basic fields on the updated account
	// based on what we already know about the requester.
	updatedAcct.CreatedAt = requestingAcct.CreatedAt
	updatedAcct.ID = requestingAcct.ID
	updatedAcct.Language = requestingAcct.Language
	updatedAcct.AvatarMediaAttachmentID = requestingAcct.AvatarMediaAttachmentID
	updatedAcct.HeaderMediaAttachmentID = requestingAcct.HeaderMediaAttachmentID

	// Pass to the processor for further updating of eg., avatar/header,
	// emojis, etc. The actual db insert/update will take place there.
	f.state.Workers.EnqueueFederator(ctx, messages.FromFederator{
		APObjectType:     ap.ObjectProfile,
		APActivityType:   ap.ActivityUpdate,
		GTSModel:         updatedAcct,
		APObjectModel:    accountable,
		ReceivingAccount: receivingAcct,
	})

	return nil
}
