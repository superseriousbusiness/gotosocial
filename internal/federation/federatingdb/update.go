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

package federatingdb

import (
	"context"
	"errors"
	"fmt"

	"codeberg.org/gruf/go-kv"
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
	l := log.Entry{}

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
		// If the receiving account wasn't set on the context, that means this request didn't pass
		// through the API, but came from inside GtS as the result of another activity on this instance. That being so,
		// we can safely just ignore this activity, since we know we've already processed it elsewhere.
		return nil
	}

	requestingAcctI := ctx.Value(ap.ContextRequestingAccount)
	if requestingAcctI == nil {
		l.Error("UPDATE: requesting account wasn't set on context")
	}
	requestingAcct, ok := requestingAcctI.(*gtsmodel.Account)
	if !ok {
		l.Error("UPDATE: requesting account was set on context but couldn't be parsed")
	}

	typeName := asType.GetTypeName()
	if typeName == ap.ActorApplication ||
		typeName == ap.ActorGroup ||
		typeName == ap.ActorOrganization ||
		typeName == ap.ActorPerson ||
		typeName == ap.ActorService {
		// it's an UPDATE to some kind of account
		var accountable ap.Accountable
		switch typeName {
		case ap.ActorApplication:
			l.Debug("got update for APPLICATION")
			i, ok := asType.(vocab.ActivityStreamsApplication)
			if !ok {
				return errors.New("UPDATE: could not convert type to application")
			}
			accountable = i
		case ap.ActorGroup:
			l.Debug("got update for GROUP")
			i, ok := asType.(vocab.ActivityStreamsGroup)
			if !ok {
				return errors.New("UPDATE: could not convert type to group")
			}
			accountable = i
		case ap.ActorOrganization:
			l.Debug("got update for ORGANIZATION")
			i, ok := asType.(vocab.ActivityStreamsOrganization)
			if !ok {
				return errors.New("UPDATE: could not convert type to organization")
			}
			accountable = i
		case ap.ActorPerson:
			l.Debug("got update for PERSON")
			i, ok := asType.(vocab.ActivityStreamsPerson)
			if !ok {
				return errors.New("UPDATE: could not convert type to person")
			}
			accountable = i
		case ap.ActorService:
			l.Debug("got update for SERVICE")
			i, ok := asType.(vocab.ActivityStreamsService)
			if !ok {
				return errors.New("UPDATE: could not convert type to service")
			}
			accountable = i
		}

		updatedAcct, err := f.typeConverter.ASRepresentationToAccount(ctx, accountable, "", true)
		if err != nil {
			return fmt.Errorf("UPDATE: error converting to account: %s", err)
		}

		if updatedAcct.Domain == config.GetHost() || updatedAcct.Domain == config.GetAccountDomain() {
			// no need to update local accounts
			// in fact, if we do this will break the shit out of things so do NOT
			return nil
		}

		if requestingAcct.URI != updatedAcct.URI {
			return fmt.Errorf("UPDATE: update for account %s was requested by account %s, this is not valid", updatedAcct.URI, requestingAcct.URI)
		}

		// set some fields here on the updatedAccount representation so we don't run into db issues
		updatedAcct.CreatedAt = requestingAcct.CreatedAt
		updatedAcct.ID = requestingAcct.ID
		updatedAcct.Language = requestingAcct.Language

		// pass to the processor for further updating of eg., avatar/header, emojis
		// the actual db insert/update will take place a bit later
		f.fedWorker.Queue(messages.FromFederator{
			APObjectType:     ap.ObjectProfile,
			APActivityType:   ap.ActivityUpdate,
			GTSModel:         updatedAcct,
			ReceivingAccount: receivingAccount,
		})
	} else if typeName == ap.ObjectNote {
		// A note is being updated
		l.Debug("got update for NOTE")
		note, ok := asType.(vocab.ActivityStreamsNote)
		if !ok {
			return errors.New("UPDATE: could not convert type to note")
		}
		if err := f.updateNote(ctx, note, receivingAccount, requestingAcct); err != nil {
			return fmt.Errorf("UPDATE: error updating note: %s", err)
		}
	}

	return nil
}

/*
	UPDATE HANDLERS
*/

// updateNote handles a Update activity with a Note type.
func (f *federatingDB) updateNote(ctx context.Context, note vocab.ActivityStreamsNote, receivingAccount *gtsmodel.Account, requestingAccount *gtsmodel.Account) error {
	l := log.WithFields(kv.Fields{
		{"receivingAccount", receivingAccount.URI},
		{"requestingAccount", requestingAccount.URI},
	}...)

	// Check if we have a forward.
	// In other words, was the note posted to our inbox by at least one actor who actually created the note, or are they just forwarding it?
	forward := true

	// note should have an attributedTo
	noteAttributedTo := note.GetActivityStreamsAttributedTo()
	if noteAttributedTo == nil {
		return errors.New("updateNote: note had no attributedTo")
	}

	// compare the attributedTo(s) with the actor who posted this to our inbox
	for attributedToIter := noteAttributedTo.Begin(); attributedToIter != noteAttributedTo.End(); attributedToIter = attributedToIter.Next() {
		if !attributedToIter.IsIRI() {
			continue
		}
		iri := attributedToIter.GetIRI()
		if requestingAccount.URI == iri.String() {
			// at least one creator of the note, and the actor who posted the note to our inbox, are the same, so it's not a forward
			forward = false
		}
	}

	// If we do have a forward, we should ignore the content for now and just dereference based on the URL/ID of the note instead, to get the note straight from the horse's mouth
	if forward {
		l.Trace("note is a forward")
		id := note.GetJSONLDId()
		if !id.IsIRI() {
			// if the note id isn't an IRI, there's nothing we can do here
			return nil
		}
		// pass the note iri into the processor and have it do the dereferencing instead of doing it here
		f.fedWorker.Queue(messages.FromFederator{
			APObjectType:     ap.ObjectNote,
			APActivityType:   ap.ActivityUpdate,
			APIri:            id.GetIRI(),
			GTSModel:         nil,
			ReceivingAccount: receivingAccount,
		})
		return nil
	}

	// if we reach this point, we know it's not a forwarded status, so proceed with processing it as normal

	status, err := f.typeConverter.ASStatusToStatus(ctx, note)
	if err != nil {
		return fmt.Errorf("updateNote: error converting note to status: %s", err)
	}

	// get the URI of the note
	noteURI := note.GetJSONLDId().GetIRI()

	// get the existing status
	existingStatus, err := f.db.GetStatusByURI(ctx, noteURI.String())
	if err != nil {
		return err
	}

	// Set values not passed by the updated status
	status.ID = existingStatus.ID
	status.Emojis = existingStatus.Emojis
	status.Local = existingStatus.Local

	if err := f.db.UpdateStatus(ctx, status); err != nil {
		return fmt.Errorf("updateNote: database error updating status: %s", err)
	}

	f.fedWorker.Queue(messages.FromFederator{
		APObjectType:     ap.ObjectNote,
		APActivityType:   ap.ActivityUpdate,
		GTSModel:         status,
		ReceivingAccount: receivingAccount,
	})

	return nil
}
