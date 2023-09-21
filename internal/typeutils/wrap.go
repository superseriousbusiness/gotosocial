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

package typeutils

import (
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// WrapPersonInUpdate ...
func (c *Converter) WrapPersonInUpdate(person vocab.ActivityStreamsPerson, originAccount *gtsmodel.Account) (vocab.ActivityStreamsUpdate, error) {
	update := streams.NewActivityStreamsUpdate()

	// set the actor
	actorURI, err := url.Parse(originAccount.URI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.URI, err)
	}
	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(actorURI)
	update.SetActivityStreamsActor(actorProp)

	// set the ID

	newID, err := id.NewRandomULID()
	if err != nil {
		return nil, err
	}

	idString := uris.GenerateURIForUpdate(originAccount.Username, newID)
	idURI, err := url.Parse(idString)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", idString, err)
	}
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(idURI)
	update.SetJSONLDId(idProp)

	// set the person as the object here
	objectProp := streams.NewActivityStreamsObjectProperty()
	objectProp.AppendActivityStreamsPerson(person)
	update.SetActivityStreamsObject(objectProp)

	// to should be public
	toURI, err := url.Parse(pub.PublicActivityPubIRI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", pub.PublicActivityPubIRI, err)
	}
	toProp := streams.NewActivityStreamsToProperty()
	toProp.AppendIRI(toURI)
	update.SetActivityStreamsTo(toProp)

	// bcc followers
	followersURI, err := url.Parse(originAccount.FollowersURI)
	if err != nil {
		return nil, gtserror.Newf("error parsing url %s: %w", originAccount.FollowersURI, err)
	}
	bccProp := streams.NewActivityStreamsBccProperty()
	bccProp.AppendIRI(followersURI)
	update.SetActivityStreamsBcc(bccProp)

	return update, nil
}

// WrapNoteInCreate wraps a Note with a Create activity.
//
// If objectIRIOnly is set to true, then the function won't put the *entire* note in the Object field of the Create,
// but just the AP URI of the note. This is useful in cases where you want to give a remote server something to dereference,
// and still have control over whether or not they're allowed to actually see the contents.
func (c *Converter) WrapNoteInCreate(note vocab.ActivityStreamsNote, objectIRIOnly bool) (vocab.ActivityStreamsCreate, error) {
	create := streams.NewActivityStreamsCreate()

	// Object property
	objectProp := streams.NewActivityStreamsObjectProperty()
	if objectIRIOnly {
		objectProp.AppendIRI(note.GetJSONLDId().GetIRI())
	} else {
		objectProp.AppendActivityStreamsNote(note)
	}
	create.SetActivityStreamsObject(objectProp)

	// ID property
	idProp := streams.NewJSONLDIdProperty()
	createID := note.GetJSONLDId().GetIRI().String() + "/activity"
	createIDIRI, err := url.Parse(createID)
	if err != nil {
		return nil, err
	}
	idProp.SetIRI(createIDIRI)
	create.SetJSONLDId(idProp)

	// Actor Property
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI, err := ap.ExtractAttributedToURI(note)
	if err != nil {
		return nil, gtserror.Newf("couldn't extract AttributedTo: %w", err)
	}
	actorProp.AppendIRI(actorIRI)
	create.SetActivityStreamsActor(actorProp)

	// Published Property
	publishedProp := streams.NewActivityStreamsPublishedProperty()
	published, err := ap.ExtractPublished(note)
	if err != nil {
		return nil, gtserror.Newf("couldn't extract Published: %w", err)
	}
	publishedProp.Set(published)
	create.SetActivityStreamsPublished(publishedProp)

	// To Property
	toProp := streams.NewActivityStreamsToProperty()
	if toURIs := ap.ExtractToURIs(note); len(toURIs) != 0 {
		for _, toURI := range toURIs {
			toProp.AppendIRI(toURI)
		}
		create.SetActivityStreamsTo(toProp)
	}

	// Cc Property
	ccProp := streams.NewActivityStreamsCcProperty()
	if ccURIs := ap.ExtractCcURIs(note); len(ccURIs) != 0 {
		for _, ccURI := range ccURIs {
			ccProp.AppendIRI(ccURI)
		}
		create.SetActivityStreamsCc(ccProp)
	}

	return create, nil
}
